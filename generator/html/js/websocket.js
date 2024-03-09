import * as THREE from 'three';
import { CSS2DObject } from 'three/addons/renderers/CSS2DRenderer.js';
import { BinaryReader } from './binary_reader.js';
import { BinaryWriter } from './binary_writer.js';

export const RepresentationType = {
    Player: 0,
    LeftHand: 1,
    RightHand: 2
}

export const ServerMessageType = {
    SetClientIDMessage: 0 + 128,
    RoomStateUpdateMessage: 1 + 128,
    RefrershGeneratorMessage: 2 + 128,
    BroadcastMessage: 3 + 128,
}


export const ClientMessageType = {
    SetOrientationMessage: 0,
    SetDisplayNameMessage: 1,
    SetSceneMessage: 2,
    SetPointerMessage: 3,
    RemovePointerMessage: 4,
}

export class WebSocketRepresentationManager {
    constructor() {
        this.rerpresentations = new Map();
    }

    AddRepresentation(key, threeObj) {
        if (this.rerpresentations.has(key)) {
            console.warn("representation manager already has a rep for key: " + key)
            // throw new Error("representation manager already has a rep for key: " + key)
        }
        this.rerpresentations.set(key, threeObj)
    }

    RemoveRepresentation(key, threeObj) {
        if (!this.rerpresentations.has(key)) {
            throw new Error("representation manager does not have a rep for key: " + key)
        }
        this.rerpresentations.set(key, threeObj)
    }

    ToMessage() {
        const writer = new BinaryWriter(true);
        writer.byte(ClientMessageType.SetOrientationMessage);

        let pos = new THREE.Vector3();
        let rot = new THREE.Quaternion();

        this.rerpresentations.forEach((threeObj, key) => {
            if (!threeObj) {
                return;
            }

            if (!threeObj.matrixWorld) {
                return;
            }

            // Somethings wrong with the matrix world, let's not include this
            // representation
            const eles = threeObj.matrixWorld.elements;
            for (let i = 0; i < eles.length; i++) {
                if (isNaN(eles[i]) || !isFinite(eles[i])) {
                    return;
                }
            }

            try {
                const worldMatrix = threeObj.matrixWorld;
                pos.setFromMatrixPosition(worldMatrix)
                rot.setFromRotationMatrix(worldMatrix)

                writer.byte(key);
                writer.float32(pos.x);
                writer.float32(pos.y);
                writer.float32(pos.z);
                writer.float32(rot.x);
                writer.float32(rot.y);
                writer.float32(rot.z);
                writer.float32(rot.w);
            } catch (error) {
                console.error(error);
                // Expected output: ReferenceError: nonExistentFunction is not defined
                // (Note: the exact output may be browser-dependent)
            }
        });
        return writer.buffer();
    }
}

export class WebSocketManager {
    constructor(
        representationManager,
        scene,
        playerConfiguration,
        clock,
        viewportSettings,
        schemaManager
    ) {
        this.representationManager = representationManager;
        this.scene = scene;
        this.playerConfiguration = playerConfiguration;
        this.viewportSettings = viewportSettings;
        this.clock = clock;
        this.schemaManager = schemaManager;

        this.lastUpdatedModel = -1;
        this.connectedPlayers = {};
        this.clientID = null;
    }

    canConnect() {
        if (window["WebSocket"]) {
            return true;
        }
        return false;
    }

    connect() {
        let websocketProtocol = "ws://";
        if (location.protocol === 'https:') {
            websocketProtocol = "wss://";
        }

        this.conn = new WebSocket(websocketProtocol + document.location.host + "/live");
        this.conn.binaryType = "arraybuffer";
        this.conn.onclose = this.onClose.bind(this);
        this.conn.onmessage = this.onMessage.bind(this);

        setInterval(this.updateServerWithCameraData.bind(this), 200);
        setInterval(this.updateServerWithSceneData.bind(this), 200);
    }

    updateServerWithSceneData() {
        if (this.viewportSettings.SettingsHaveChanged() === false) {
            return;
        }

        const settings = this.viewportSettings.GetFolder();

        const writer = new BinaryWriter(true);
        writer.byte(ClientMessageType.SetSceneMessage);
        writer.bool(settings.renderWireframe);
        writer.bool(settings.antiAlias);
        writer.bool(settings.xrEnabled);

        const color = new THREE.Color();
        color.set(settings.fog.color);
        writer.byte(color.r * 255);
        writer.byte(color.g * 255);
        writer.byte(color.b * 255);
        writer.byte(color.a * 255);
        writer.float32(settings.fog.near);
        writer.float32(settings.fog.far);

        color.set(settings.background);
        writer.byte(color.r * 255);
        writer.byte(color.g * 255);
        writer.byte(color.b * 255);
        writer.byte(color.a * 255);

        color.set(settings.lighting);
        writer.byte(color.r * 255);
        writer.byte(color.g * 255);
        writer.byte(color.b * 255);
        writer.byte(color.a * 255);

        color.set(settings.ground);
        writer.byte(color.r * 255);
        writer.byte(color.g * 255);
        writer.byte(color.b * 255);
        writer.byte(color.a * 255);

        this.conn.send(writer.buffer());

        this.viewportSettings.ResetSettingsHaveChanged();
    }

    onClose(evt) {
        console.log("connection closed", evt)
    }

    createPlayerObject(name, playerData) {
        const newPlayer = new THREE.Group();
        newPlayer.name = "player";

        const sphere = new THREE.Mesh(
            this.playerConfiguration.playerGeometry,
            this.playerConfiguration.playerMaterial
        );
        sphere.position.z += 0.5;
        newPlayer.add(sphere);

        const eyeSize = 0.15;
        const eyeSpacing = 0.3;

        const leftEye = new THREE.Mesh(
            this.playerConfiguration.playerGeometry,
            this.playerConfiguration.playerEyeMaterial
        );
        leftEye.scale.x = eyeSize;
        leftEye.scale.y = eyeSize;
        leftEye.scale.z = eyeSize;
        leftEye.position.x = eyeSpacing;
        leftEye.position.z = - 0.5;
        leftEye.position.y = + 0.25;
        newPlayer.add(leftEye);

        const rightEye = new THREE.Mesh(
            this.playerConfiguration.playerGeometry,
            this.playerConfiguration.playerEyeMaterial
        );
        rightEye.scale.x = eyeSize;
        rightEye.scale.y = eyeSize;
        rightEye.scale.z = eyeSize;
        rightEye.position.x = - eyeSpacing;
        rightEye.position.z = - 0.5;
        rightEye.position.y = + 0.25;
        newPlayer.add(rightEye);

        const text = document.createElement('div');
        text.className = 'label';
        text.style.color = '#000000';
        text.textContent = name;
        text.style.fontSize = "30px";

        const label = new CSS2DObject(text);
        label.position.y += 0.75;
        newPlayer.add(label);


        newPlayer.position.x = playerData.position.x;
        newPlayer.position.y = playerData.position.y;
        newPlayer.position.z = playerData.position.z;

        newPlayer.scale.set(0.25, 0.25, 0.25)
        this.scene.add(newPlayer);

        return {
            desiredPosition: playerData.position,
            desiredRotation: playerData.rotation,
            obj: newPlayer,
            cleanup: () => {
                newPlayer.remove(label);
                this.scene.remove(newPlayer);
            }
        };
    }

    createHandObject(handData) {
        console.log("creating hand")
        const hand = new THREE.Group();
        hand.name = "hand";

        const sphere = new THREE.Mesh(
            this.playerConfiguration.playerGeometry,
            this.playerConfiguration.playerMaterial
        );
        sphere.scale.set(0.1, 0.1, 0.1)
        hand.add(sphere);

        hand.position.x = handData.position.x;
        hand.position.y = handData.position.y;
        hand.position.z = handData.position.z;

        this.scene.add(hand);

        return {
            desiredPosition: handData.position,
            desiredRotation: handData.rotation,
            obj: hand,
            cleanup: () => {
                this.scene.remove(hand);
            }
        };
    }


    setupPlayer(key, playerData) {
        const resps = []

        if (playerData.representation) {
            playerData.representation.forEach((rep) => {
                switch (rep.type) {
                    case RepresentationType.Player:
                        resps.push(this.createPlayerObject(playerData.name, rep))
                        break;

                    case RepresentationType.LeftHand:
                        resps.push(this.createHandObject(rep));
                        break;

                    case RepresentationType.RightHand:
                        resps.push(this.createHandObject(rep));
                        break;

                    default:
                        break;
                }
            })
        }


        this.connectedPlayers[key] = {
            representation: resps,
        };
    }

    update() {
        const delta = this.clock.getDelta();

        for (const [key, player] of Object.entries(this.connectedPlayers)) {
            if (!player.representation) {
                continue;
            }

            for (let i = 0; i < player.representation.length; i++) {
                const rep = player.representation[i];
                const dr = rep.desiredRotation;

                const q = new THREE.Quaternion(dr.x, dr.y, dr.z, dr.w);
                if (!rep.obj.quaternion.equals(q)) {
                    rep.obj.quaternion.rotateTowards(q, delta * 2);
                }

                const pp = rep.obj.position;
                const dp = rep.desiredPosition;

                pp.x = pp.x + ((dp.x - pp.x) * delta * 4);
                pp.y = pp.y + ((dp.y - pp.y) * delta * 4);
                pp.z = pp.z + ((dp.z - pp.z) * delta * 4);
            }
        }
    }

    onSetClientID(messageData) {
        console.log(messageData, "test");
        this.clientID = messageData;
    }

    onRoomStateUpdate(messageData) {
        if (this.lastUpdatedModel !== messageData.ModelVersion) {
            this.lastUpdatedModel = messageData.ModelVersion;
            this.schemaManager.refreshSchema();
        }

        if (this.viewportSettings.SettingsHaveChanged() === false) {
            const webScene = messageData.WebScene;

            for (const [setting, data] of Object.entries(webScene)) {
                this.viewportSettings.UpdateSetting(setting, data)
            }

            for (const [setting, data] of Object.entries(webScene.fog)) {
                this.viewportSettings.UpdateSetting("fog/" + setting, data)
            }
        }

        const playersUpdated = {}
        for (const [key, value] of Object.entries(this.connectedPlayers)) {
            playersUpdated[key] = false;
        }

        for (const [playerID, serverPlayer] of Object.entries(messageData.Players)) {
            if (serverPlayer == null) {
                continue;
            }

            // We don't want to create a representation of ourselves
            if (playerID == this.clientID) {
                continue;
            }

            if (!serverPlayer.representation) {
                continue;
            }

            playersUpdated[playerID] = true;

            if (playerID in this.connectedPlayers && this.connectedPlayers[playerID].representation.length === serverPlayer.representation.length) {
                const player = this.connectedPlayers[playerID];
                for (let i = 0; i < player.representation.length; i++) {
                    console.log("updating " + serverPlayer.representation[i].type)
                    player.representation[i].desiredPosition.x = serverPlayer.representation[i].position.x;
                    player.representation[i].desiredPosition.y = serverPlayer.representation[i].position.y;
                    player.representation[i].desiredPosition.z = serverPlayer.representation[i].position.z;

                    player.representation[i].desiredRotation.x = serverPlayer.representation[i].rotation.x;
                    player.representation[i].desiredRotation.y = serverPlayer.representation[i].rotation.y;
                    player.representation[i].desiredRotation.z = serverPlayer.representation[i].rotation.z;
                    player.representation[i].desiredRotation.w = serverPlayer.representation[i].rotation.w;
                }

            } else {
                if (playerID in this.connectedPlayers) {
                    this.removePlayer(playerID);
                }

                // Create a new Player!
                this.setupPlayer(playerID, serverPlayer);
            }
        }

        // Remove all players that weren't contained within the update
        for (const [playerID, updated] of Object.entries(playersUpdated)) {
            if (updated) {
                continue;
            }
            this.removePlayer(playerID);
        }
    }

    removePlayer(playerID) {
        this.connectedPlayers[playerID].representation.forEach(rep => rep.cleanup());
        delete this.connectedPlayers[playerID];
    }

    onMessage(evt) {
        const dataView = new DataView(evt.data);
        dataView.byteLength
        const reader = new BinaryReader(dataView);

        const messageType = reader.Byte();
        switch (messageType) {
            case ServerMessageType.SetClientIDMessage:
                this.onSetClientID(reader.String(reader.RemainingLength()));
                break;

            case ServerMessageType.RoomStateUpdateMessage:
                const room = {
                    ModelVersion: 0,
                    WebScene: {
                        renderWireframe: false,
                        antiAlias: false,
                        xrEnabled: false,
                        fog: {
                            color: "",
                            near: 5,
                            far: 25
                        },
                        background: "",
                        lighting: "",
                        ground: "",
                    },
                    Players: {}
                }

                room.ModelVersion = reader.UInt32();
                room.WebScene.renderWireframe = reader.Bool();
                room.WebScene.antiAlias = reader.Bool();
                room.WebScene.xrEnabled = reader.Bool();

                const color = new THREE.Color();

                color.r = reader.Byte() / 255;
                color.g = reader.Byte() / 255;
                color.b = reader.Byte() / 255;
                color.a = reader.Byte() / 255;
                room.WebScene.fog.color = "#" + color.getHexString();
                room.WebScene.fog.near = reader.Float32();
                room.WebScene.fog.far = reader.Float32();

                color.r = reader.Byte() / 255;
                color.g = reader.Byte() / 255;
                color.b = reader.Byte() / 255;
                color.a = reader.Byte() / 255;
                room.WebScene.background = "#" + color.getHexString();

                color.r = reader.Byte() / 255;
                color.g = reader.Byte() / 255;
                color.b = reader.Byte() / 255;
                color.a = reader.Byte() / 255;
                room.WebScene.lighting = "#" + color.getHexString();

                color.r = reader.Byte() / 255;
                color.g = reader.Byte() / 255;
                color.b = reader.Byte() / 255;
                color.a = reader.Byte() / 255;
                room.WebScene.ground = "#" + color.getHexString();

                const numPlayers = reader.Byte();
                for (let playerIndex = 0; playerIndex < numPlayers; playerIndex++) {
                    const id = reader.String(reader.Byte());
                    const name = reader.String(reader.Byte());

                    const representations = [];
                    const repLen = reader.Byte();
                    for (let repI = 0; repI < repLen; repI++) {
                        const rep = {
                            type: 0,
                            position: {
                                x: 0,
                                y: 0,
                                z: 0
                            },
                            rotation: {
                                x: 0,
                                y: 0,
                                z: 0,
                                w: 0
                            }
                        }

                        rep.type = reader.Byte();
                        rep.position.x = reader.Float32();
                        rep.position.y = reader.Float32();
                        rep.position.z = reader.Float32();
                        rep.rotation.x = reader.Float32();
                        rep.rotation.y = reader.Float32();
                        rep.rotation.z = reader.Float32();
                        rep.rotation.w = reader.Float32();

                        representations.push(rep);
                    }

                    room.Players[id] = {
                        name: name,
                        representation: representations
                    }
                }

                this.onRoomStateUpdate(room);
                break;

            case ServerMessageType.RefrershGeneratorMessage:
                break;

            case ServerMessageType.BroadcastMessage:
                break;
        }
    };

    updateServerWithCameraData() {
        if (this.conn.readyState !== this.conn.OPEN) {
            return;
        }
        this.conn.send(this.representationManager.ToMessage());
    }
}