import fs from 'fs';
import { nativeTheme } from 'electron';
import { SupportedLanguages } from '../shared/languages';
import log from '../shared/log';
import { mergeDeep } from '../shared/object';
import {
  GetSettingsRequest,
  GetSettingsResponse,
  UpdateSettingsRequest,
  UpdateSettingsResponse,
} from '../shared/rpc';
import { LanguageSettings, Settings } from '../shared/settings';
import { ensureFile } from './fs';
import { RPCHandler } from './rpc';

export class DesktopSettings extends Settings {
  constructor(
    file: string,
    id?: string,
    lastProject?: string,
    languages?: Record<SupportedLanguages, LanguageSettings>,
    stdoutMaxSize?: number
  ) {
    super(file, nativeTheme.shouldUseDarkColors ? 'dark' : 'light', id, lastProject, languages, stdoutMaxSize)
  }
  static fromFile(settingsFile: string) {
    let existingSettingsString: Buffer | null = null;
    try {
      existingSettingsString = fs.readFileSync(settingsFile);
    } catch (e) {
      // Fine if it doesn't exist
    }

    let existingSettings: Partial<Settings> = {
      file: settingsFile,
    };
    let s: string = '';
    if (existingSettingsString && (s = existingSettingsString.toString())) {
      try {
        existingSettings = JSON.parse(s);
        // Migrate from .uid to .id
        if ((existingSettings as any).uid) {
          existingSettings.id = (existingSettings as any).uid;
          delete (existingSettings as any).uid;
        }
      } catch (e) {
        const backupFile = settingsFile + '.bak';
        log.error(
          `Settings file has been corrupted, renaming to ${backupFile}`,
          e
        );
        fs.renameSync(settingsFile, backupFile);
      }
    }

    return mergeDeep(new DesktopSettings(settingsFile), existingSettings);
  }

  save() {
    return fs.writeFileSync(this.file, JSON.stringify(this));
  }

  getUpdateHandler(): RPCHandler<
    UpdateSettingsRequest,
    UpdateSettingsResponse
  > {
    return {
      resource: 'updateSettings',
      handler: async (_: string, settings: Settings) => {
        Object.assign(this, settings);
        return this.save();
      },
    };
  }

  getGetHandler(): RPCHandler<GetSettingsRequest, GetSettingsResponse> {
    return {
      resource: 'getSettings',
      handler: async () => {
        Object.assign(this, DesktopSettings.fromFile(this.file));
        return this;
      },
    };
  }
}

export let SETTINGS = new DesktopSettings('');

export function loadSettings(settingsFile?: string): DesktopSettings {
  if (!settingsFile) {
    settingsFile = '.settings';
  }
  const fullName = ensureFile(settingsFile);
  SETTINGS = DesktopSettings.fromFile(fullName);
  return SETTINGS;
}
