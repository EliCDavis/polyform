import * as THREE from 'three';
import { CSS2DObject } from 'three/addons/renderers/CSS2DRenderer.js';

export class WebSocketRepresentationManager {
    constructor() {
        this.rerpresentations = new Map();
    }

    addRepresentation(key, threeObj) {
        if (this.rerpresentations.has(key)) {
            throw new Error("representation manager already has a rep for key: " + key)
        }
        this.rerpresentations.set(key, threeObj)
    }

    removeRepresentation(key, threeObj) {
        if (this.rerpresentations.has(key)) {
            throw new Error("representation manager already has a rep for key: " + key)
        }
        this.rerpresentations.set(key, threeObj)
    }

    toMessage() {
        const message = [];
        this.rerpresentations.forEach((threeObj, key) => {
            message.push({
                type: key,
                "position": {
                    "x": threeObj.position.x,
                    "y": threeObj.position.y,
                    "z": threeObj.position.z,
                },
                "rotation": {
                    "x": threeObj.quaternion.x,
                    "y": threeObj.quaternion.y,
                    "z": threeObj.quaternion.z,
                    "w": threeObj.quaternion.w,
                }
            })
        });
        return message;
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
        this.conn.onclose = this.onClose.bind(this);
        this.conn.onmessage = this.onMessage.bind(this);

        setInterval(this.updateServerWithCameraData.bind(this), 200);
        setInterval(this.updateServerWithSceneData.bind(this), 200);
    }

    updateServerWithSceneData() {
        if (this.viewportSettings.SettingsHaveChanged() === false) {
            return;
        }

        this.conn.send(JSON.stringify({
            "type": "Client-SetScene",
            "data": this.viewportSettings.GetFolder()
        }));

        this.viewportSettings.ResetSettingsHaveChanged();
    }

    onClose(evt) {
        console.log("connection closed", evt)
    }

    createPlayerObject(name, playerData) {
        const newPlayer = new THREE.Group();

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

    setupPlayer(key, playerData) {
        const resps = []

        console.log(playerData);

        if (playerData.representation) {
            playerData.representation.forEach((rep) => {
                switch (rep.type) {
                    case "player":
                        resps.push(this.createPlayerObject(playerData.name, rep))
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
                    rep.obj.quaternion.rotateTowards(q, delta *2);
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
                console.log("updating...")
                const player = this.connectedPlayers[playerID];
                for (let i = 0; i < player.representation.length; i++) {
                    player.representation[i].desiredPosition.x = serverPlayer.representation[i].position.x;
                    player.representation[i].desiredPosition.y = serverPlayer.representation[i].position.y;
                    player.representation[i].desiredPosition.z = serverPlayer.representation[i].position.z;

                    player.representation[i].desiredRotation.x = serverPlayer.representation[i].rotation.x;
                    player.representation[i].desiredRotation.y = serverPlayer.representation[i].rotation.y;
                    player.representation[i].desiredRotation.z = serverPlayer.representation[i].rotation.z;
                    player.representation[i].desiredRotation.w = serverPlayer.representation[i].rotation.w;
                }

            } else {
                // Create a new Player!
                this.setupPlayer(playerID, serverPlayer);
            }
        }

        // Remove all players that weren't contained within the update
        for (const [playerID, updated] of Object.entries(playersUpdated)) {
            if (updated) {
                continue;
            }

            this.connectedPlayers[playerID].representation.forEach(rep => rep.cleanup());
            delete this.connectedPlayers[playerID];
        }
    }

    onMessage(evt) {
        const message = JSON.parse(evt.data);

        switch (message.type) {
            case "Server-SetClientID":
                this.onSetClientID(message.data)
                break;

            case "Server-RoomStateUpdate":
                this.onRoomStateUpdate(message.data);
                break;

            case "Server-RefreshGenerator":
                break;

            case "Server-Broadcast":
                break;
        }
    };

    updateServerWithCameraData() {
        if (this.conn.readyState !== this.conn.OPEN) {
            return;
        }
        this.conn.send(JSON.stringify({
            "type": "Client-SetOrientation",
            "data": {
                "representation": this.representationManager.toMessage(),
            }
        }));
    }
}
