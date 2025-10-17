const BackendAddr = "http://localhost:8080";
const WebsocketAddr = BackendAddr.startsWith("https://") ?
    BackendAddr.replace("https://", "wss://") :
    BackendAddr.replace("http://", "ws://");

const backend = {
    /** @type {WebSocket} */
    conn: null,
    /**
     * Create a new dot on the backend.
     * @param {string} id - Unique ID of the dot
     * @param {Vec3} position - Position of the dot
     * @param {Vec3} color - Color of the dot
     */
    createDot: async function (id, position, color) {
        const body = JSON.stringify({ id, position, color });
        const response = await fetch(BackendAddr + "/api/dots", { method: "POST", body });
        if (!response.ok)
            throw new Error("response has non-2xx status code: " + response.status.toString());
    },
    /**
     * Get the list of all dots from the backend.
     * @returns {Promise<Dot[]>}
     */
    listDots: async function () {
        const response = await fetch(BackendAddr + "/api/dots");
        if (!response.ok)
            throw new Error("response has non-2xx status code: " + response.status.toString());

        /** @type {Object<string, Dot>} */
        const map = await response.json();

        return Object.values(map).map(v => ({
            ...v,
            position: new Vec3(v.position.x, v.position.y, v.position.z),
            velocity: new Vec3(v.velocity.x, v.velocity.y, v.velocity.z),
            color: new Vec3(v.color.x, v.color.y, v.color.z),
        }));
    },
    /**
     * Initialize a websocket connection with the backend.
     * @param {(message: { event: string, data: any }) => void} onMessage - Callback for received messages
     */
    initConnection: async function (onMessage) {
        backend.conn = new WebSocket(WebsocketAddr + "/api/conn");

        backend.conn.addEventListener("open", function (event) {
            console.info("Successfully established websocket connection.");
        });

        backend.conn.addEventListener("close", function (event) {
            console.info("Websocket connection closed.");
        });

        backend.conn.addEventListener("error", function (event) {
            console.info("Error in websocket connection", event);
        });

        backend.conn.addEventListener("message", function (event) {
            onMessage(JSON.parse(event.data));
        });
    }
}
