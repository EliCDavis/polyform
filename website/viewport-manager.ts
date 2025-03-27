export class ViewportSetting {

    dataKey: string

    dataHolder: any;

    subs: Array<() => void>;

    updater: () => void;

    setting: any;

    constructor(
        dataKey: string,
        dataHolder: any,
        setting,
        updater: () => void
    ) {
        this.subs = [];

        this.dataKey = dataKey;
        this.dataHolder = dataHolder;
        this.updater = updater;
        this.setting = setting.listen()
            .onChange((evt) => {
                this.updater();
                this.subs.forEach(sub => sub());
            });
    }

    AddSubscriber(sub: () => void): void {
        this.subs.push(sub);
    }

    Update(data) {
        if (this.dataHolder[this.dataKey] === data) {
            return;
        }
        this.dataHolder[this.dataKey] = data;
        this.setting.updateDisplay();
        this.updater();
    }
}

export class ViewportManager {

    settings: Map<String, ViewportSetting>;

    folder: any

    viewportSettingsChanged: boolean

    constructor(folder: any) {
        this.settings = new Map<String, ViewportSetting>();
        this.viewportSettingsChanged = false;
        this.folder = folder;
    }

    GetFolder(): any {
        return this.folder;
    }

    AddSetting(id, setting) {
        if (this.settings.has(id)) {
            throw new Error("Viewport Manager already has setting with id '" + id + "' registered");
        }
        this.settings.set(id, setting);
        setting.AddSubscriber(this.onSettingChanged.bind(this));
    }

    onSettingChanged() {
        this.viewportSettingsChanged = true;
    }

    SettingsHaveChanged() {
        return this.viewportSettingsChanged;
    }

    ResetSettingsHaveChanged() {
        this.viewportSettingsChanged = false;
    }

    UpdateSetting(id, data) {
        if (this.settings.has(id) == false) {
            return;
        }
        this.settings.get(id).Update(data)
    }
}