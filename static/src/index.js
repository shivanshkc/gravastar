// Gravastar uses a square screen. So, width = height = resolution
const Resolution = 1000;

async function main() {
    // Get the UI elements.
    const canvas = document.getElementById("world");
    const syncButton = document.getElementById("sync-button");
    const muteButton = document.getElementById("mute-button");
    const muteButtonIcon = document.getElementById("mute-button-icon");

    // Make sure all elements are found.
    if (!canvas || !syncButton || !muteButton || !muteButtonIcon) {
        console.error("could not find all UI elements");
        return;
    }

    // Keep the canvas size same as its DOM size.
    correctCanvasSize(canvas);
    window.onresize = () => correctCanvasSize(canvas);

    // Setup beeper/sound-system without blocking.
    const beeper = new Beeper();
    beeper.loadCollisionSound()
        .then(() => console.info("Collision sound loaded."))
        .catch((err) => console.error("Failed to load collision sound:", err));

    // Mute control. This is also responsible for non-suspending the AudioContext.
    muteButton.onclick = async function(event) {
        const current = muteButtonIcon.innerText;
        muteButtonIcon.innerText = current === "volume_off" ? "volume_up" : "volume_off";
        await beeper.toggleMute();
    };

    // Initialize gravity engine.
    const engine = new GravityEngine(Resolution, Resolution);
    // Play the sound on collision.
    engine.setCollisionCallback((dot) => beeper.playCollision());

    // Attach on-click actions.
    canvas.onclick = onCanvasClick(canvas, engine);
    syncButton.onclick = onSyncClick(engine);

    // Sync state with the backend initially without blocking.
    syncButton.onclick(null);

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
    }).then(() => {});

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

        const deltaSec = (timestamp - last) / 1000;
        // Basic defense against large deltas. These are observed when users switch tabs.
        if (deltaSec < 0.1) engine.tick(deltaSec);
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

    // Keep the opacity settings of the instruction local to itself.
    ctx.save();
    drawInstruction(canvas, ctx);
    ctx.restore();

    engine.getDots().forEach(dot => {
        // Scale position and color.
        const posX = dot.position.x * canvas.width / Resolution;
        const posY = dot.position.y * canvas.height / Resolution;
        const color = dot.color.mul(255);

        // Draw trail.
        if (dot.trail && dot.trail.length > 1) {
            for (let i = 0; i < dot.trail.length - 1; i++) {
                const trailPosX = dot.trail[i].x * canvas.width / Resolution;
                const trailPosY = dot.trail[i].y * canvas.height / Resolution;

                // Calculate opacity based on position in trail (newer = more opaque).
                const opacity = (i + 1) / dot.trail.length * 0.5;

                ctx.beginPath();
                ctx.arc(trailPosX, trailPosY, dot.radius * 0.4, 0, 2 * Math.PI);
                ctx.fillStyle = `rgba(${color.x}, ${color.y}, ${color.z}, ${opacity})`;
                ctx.fill();
            }
        }

        // Draw the dot.
        ctx.beginPath();
        ctx.arc(posX, posY, dot.radius, 0, 2 * Math.PI);
        ctx.fillStyle = `rgb(${color.x}, ${color.y}, ${color.z})`;
        ctx.fill();

        // Skip highlight boundary if it is someone else's dot.
        if (!localStorage.getItem("dot-"+dot.id)) return;

        // Draw highlight boundary.
        ctx.beginPath();
        ctx.arc(posX, posY, dot.radius + 5, 0, 2 * Math.PI);
        ctx.strokeStyle = "white";
        ctx.lineWidth = 2;
        ctx.stroke();
    });
}

// Parameters required to render the instruction text.
drawInstruction.textOpacity = 1.0;
drawInstruction.firstCallTimestamp = -1;
drawInstruction.maxRenderTime = 3000;

function drawInstruction(canvas, ctx) {
    if (drawInstruction.firstCallTimestamp === -1) drawInstruction.firstCallTimestamp = Date.now();

    // No write operations on these three quantities below this point. So, destructuring can be used for brevity.
    const { firstCallTimestamp, textOpacity, maxRenderTime } = drawInstruction;
    // Don't draw anything beyond the maxRenderTime.
    if (Date.now() - firstCallTimestamp > maxRenderTime) return;

    // Set font and styles for text.
    ctx.font = "16px Roboto";
    ctx.fillStyle = "grey";
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";

    // Until the last second, draw with full opacity.
    if (Date.now() - firstCallTimestamp < maxRenderTime - 1000) {
        ctx.fillText("Tap on the screen to create gravitationally active bodies.", canvas.width / 2, canvas.height / 2);
        return;
    }

    // Gradually reduce the opacity during the last second.
    let opacity = ((textOpacity * maxRenderTime) - (Date.now() - firstCallTimestamp)) / 1000;
    if (opacity < 0) return;

    ctx.globalAlpha = opacity;
    ctx.fillText("Tap on the screen to create gravitationally active bodies.", canvas.width / 2, canvas.height / 2);
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
        engine.addDot({ id, mass: 1, radius: 5, position, velocity, color });
        localStorage.setItem("dot-" + id, "ok");

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
 * @returns {(event: MouseEvent) => Promise<void>}
 */
function onSyncClick(engine) {
    return async function (event) {
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
