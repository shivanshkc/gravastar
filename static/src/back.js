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
     */
    createDot: async function (id, position) {
        const body = JSON.stringify({ id, position });
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

        /** @type {{[key: string]: Dot}} */
        const body = await response.json();

        // Explicitly convert vectors to the Vec3 type to access methods.
        return Object.values(body).map(d => {
            d.trail = [];
            d.position =  new Vec3(d.position.x, d.position.y, d.position.z);
            d.velocity = new Vec3(d.velocity.x, d.velocity.y, d.velocity.z);
            return d;
        });
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
