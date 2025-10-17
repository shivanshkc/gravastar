// Gravastar uses a square screen. So, width = height = resolution
const Resolution = 1000;

async function main() {
    // Get the canvas.
    const canvas = document.getElementById("world");
    if (!canvas) {
        console.error("could not find canvas");
        return;
    }

    const syncButton = document.getElementById("sync-button");
    if (!syncButton) {
        console.error("could not find sync button");
        return;
    }

    // Keep the canvas size same as its DOM size.
    correctCanvasSize(canvas);
    window.onresize = () => correctCanvasSize(canvas);

    // Initialize gravity engine.
    const engine = new GravityEngine(Resolution, Resolution);
    // Sync with backend state non-blockingly.

    // Attach on-click actions.
    canvas.onclick = onCanvasClick(canvas, engine);
    syncButton.onclick = onSyncClick(engine);

    // Sync state with the backend initially without blocking.
    syncButton.onclick();

    // Initialize websocket connection without blocking.
    backend.initConnection(function (message) {
        switch (message.event) {
            case "DotCreated":
                /** @type {Dot} */
                const data = message.data;

                // Dot parameters.
                const position = new Vec3(data.position.x, data.position.y, 0);
                const velocity = new Vec3(data.velocity.x, data.velocity.y, data.velocity.z);
                const color = new Vec3(data.color.x, data.color.y, data.color.z);

                engine.addDot({
                    id: data.id, mass: data.mass, radius: data.radius,
                    position, velocity, color
                });
                break;
            default:
                console.warn("Unknown application event from websocket:", message.event);
                break;
        }
    });

    // Start the simulation.
    render(canvas, engine);
}

/**
 * Runs the engine every frame and draw the output on canvas.
 * @param {HTMLCanvasElement} canvas
 * @param {GravityEngine} engine
 */
function render(canvas, engine) {
    // To calculate delta.
    let last = 0;

    /**
     * Function that will be called for each frame.
     * @param {DOMHighResTimeStamp} timestamp
     */
    const update = function (timestamp) {
        // Render first.
        draw(canvas, engine);

        // Simulate.
        engine.tick((timestamp - last) / 1000);
        last = timestamp;

        // Next frame.
        requestAnimationFrame(update);
    }

    // Trigger the first frame.
    requestAnimationFrame(update);
}

/**
 * Draw the dots from the engine onto the canvas.
 * @param {HTMLCanvasElement} canvas
 * @param {GravityEngine} engine
 */
function draw(canvas, engine) {
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    // Clear the canvas.
    ctx.clearRect(0, 0, canvas.width, canvas.height);

    engine.getDots().forEach(dot => {
        // Scale position and color.
        const posX = dot.position.x * canvas.width / Resolution;
        const posY = dot.position.y * canvas.height / Resolution;
        const color = dot.color.mul(255);

        // Draw.
        ctx.beginPath();
        ctx.arc(posX, posY, dot.radius, 0, 2 * Math.PI);
        ctx.fillStyle = `rgb(${color.x}, ${color.y}, ${color.z})`;
        ctx.fill();
    });
}

/**
 * Returns the on-click handler for the canvas.
 * @param {HTMLCanvasElement} canvas
 * @param {GravityEngine} engine
 * @returns {(p: PointerEvent) => Promise<void>}
 */
function onCanvasClick(canvas, engine) {
    return async function (event) {
        const rect = canvas.getBoundingClientRect();

        // Click locations scaled as per the required resolution.
        const x = (event.clientX - rect.left) * Resolution / rect.width;
        const y = (event.clientY - rect.top) * Resolution / rect.height;

        // Dot parameters.
        const id = crypto.randomUUID();
        const position = new Vec3(x, y, 0);
        const velocity = Vec3.zero;
        const color = brightRandom();

        // Add the dot to the engine.
        engine.addDot({ id, mass: 1, radius: 3, position, velocity, color });

        try {
            // Send new dot's info to the backend.
            await backend.createDot(id, position, color);
        } catch (err) {
            console.error("error in Create Dot API:", err);
            // Do not disturb the local simulation.
        }
    };
}

/**
 * Returns the on-click handler for the sync button.
 * @param {GravityEngine} engine
 * @returns {() => Promise<void>}
 */
function onSyncClick(engine) {
    return async function () {
        try {
            const list = await backend.listDots();
            engine.setDots(list);
        } catch (err) {
            console.error("error in List Dots API:", err);
            // Do not disturb the local simulation.
        }
    }
}

/**
 * Sets the canvas dimensions equal to its DOM dimensions.
 *
 * Canvas apparently has two different sizes. One which can be set using CSS,
 * and another that is set using the "height" and "width" property of the canvas HTML element itself.
 *
 * This method sets the latter to the former.
 * @param {HTMLCanvasElement} canvas
 */
function correctCanvasSize(canvas) {
    const rect = canvas.getBoundingClientRect();
    canvas.width = rect.width;
    canvas.height = rect.height;
}

/**
 * Returns a bright random color.
 * @returns {Vec3}
 */
function brightRandom() {
    return new Vec3(
        Math.random() * 0.5 + 0.5,
        Math.random() * 0.5 + 0.5,
        Math.random() * 0.5 + 0.5,
    );
}

main();
