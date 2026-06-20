
export interface ViewportFogSettings {
    color: string,
    near: number,
    far: number,
}

export interface ViewportSettings {
    renderWireframe: boolean,
    fog: ViewportFogSettings
    background: string,
    lighting: string,
    ground: string
}
