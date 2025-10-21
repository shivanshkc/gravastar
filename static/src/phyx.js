/**
 * @typedef {Object} Dot
 * @property {string} id - The unique identifier of the dot.
 * @property {number} mass - The mass of the dot.
 * @property {number} radius - The radius of the dot.
 * @property {Vec3} position - The position vector.
 * @property {Vec3} velocity - The velocity vector.
 * @property {Vec3} color - The color vector in RGB format, range is [0-1].
 * @property {Vec3[]} trail - Array of recent positions for drawing trails.
 */

/** @type {number} */
const GravitationalConstant = 1;

/** @type {number} */
const MaxTrailLength = 100;

/**
 * A simple 2D gravity simulation engine that updates
 * dot positions and velocities under mutual attraction.
 */
class GravityEngine {
    /**
     * Creates a new gravity engine.
     * @param {number} width - The width of the simulation area.
     * @param {number} height - The height of the simulation area.
     */
    constructor(width, height) {
        /** @type {Dot[]} */
        this.dots = [];

        /** @type {Set<string>} */
        this.ids = new Set();

        /** @type {number} */
        this.width = width;

        /** @type {number} */
        this.height = height;

        /** @type {Function|null} */
        this.onCollision = null;
    }

    /**
     * Set collision callback function
     * @param {Function} callback - Function to call when collision occurs
     */
    setCollisionCallback(callback) {
        this.onCollision = callback;
    }

    /**
     * Replace all dots in the simulation.
     * @param {Dot[]} dots - Dots to set.
     */
    setDots(dots) {
        this.dots = [...dots];
        this.ids = new Set(this.dots.map(d => d.id));
    }

    /**
     * Returns a shallow copy of all current dots.
     * @returns {Dot[]} The list of current dots.
     */
    getDots() {
        return [...this.dots];
    }

    /**
     * Add a new dot to the simulation.
     * @param {Dot} dot - Dot to add.
     */
    addDot(dot) {
        if (this.ids.has(dot.id)) return;
        this.ids.add(dot.id);
        dot.trail = [];
        this.dots.push(dot);
    }

    /**
     * Advances the simulation by one timestep.
     * @param {number} deltaSec - The timestep in seconds.
     */
    tick(deltaSec) {
        if (this.dots.length === 0) return;

        for (let i = 0; i < this.dots.length; i++) {
            const thisDot = this.dots[i];
            let totalAcceleration = Vec3.zero;

            for (let j = 0; j < this.dots.length; j++) {
                if (i === j) continue;

                const otherDot = this.dots[j];

                // Calculate relative position (scaled down to avoid huge forces).
                const distance = otherDot.position.sub(thisDot.position).div(1000);
                const distanceMag = distance.mag();
                const distanceMagSquared = distanceMag * distanceMag;

                // Avoid infinite acceleration at short range.
                const softeningSquared = 0.05 * 0.05;
                const denominator = (distanceMagSquared + softeningSquared) * distanceMag;

                // Newton's Gravitation formula (vector form).
                const acceleration = distance.mul(GravitationalConstant * otherDot.mass / denominator);
                totalAcceleration = totalAcceleration.add(acceleration);
            }

            // Second law of motion to calculate displacement.
            const halfAtSquared = totalAcceleration.mul(0.5 * deltaSec * deltaSec);
            const displacement = thisDot.velocity.mul(deltaSec).add(halfAtSquared);

            // Add current position to trail before updating
            if (!thisDot.trail) thisDot.trail = [];
            thisDot.trail.push(new Vec3(thisDot.position.x, thisDot.position.y, thisDot.position.z));

            // Cap the trail length.
            if (thisDot.trail.length > MaxTrailLength) thisDot.trail.shift();

            thisDot.position = thisDot.position.add(displacement);

            // First law of motion of calculate final velocity.
            thisDot.velocity = thisDot.velocity.add(totalAcceleration.mul(deltaSec));

            // --- Wall collisions ---
            let collisionOccurred = false;

            if (thisDot.position.x - thisDot.radius <= 0) {
                thisDot.position.x = thisDot.radius;
                thisDot.velocity.x = -thisDot.velocity.x;
                collisionOccurred = true;
            }
            if (thisDot.position.x + thisDot.radius >= this.width) {
                thisDot.position.x = this.width - thisDot.radius;
                thisDot.velocity.x = -thisDot.velocity.x;
                collisionOccurred = true;
            }
            if (thisDot.position.y - thisDot.radius <= 0) {
                thisDot.position.y = thisDot.radius;
                thisDot.velocity.y = -thisDot.velocity.y;
                collisionOccurred = true;
            }
            if (thisDot.position.y + thisDot.radius >= this.height) {
                thisDot.position.y = this.height - thisDot.radius;
                thisDot.velocity.y = -thisDot.velocity.y;
                collisionOccurred = true;
            }

            // Trigger collision callback if collision occurred.
            if (collisionOccurred && this.onCollision) this.onCollision(thisDot);

            this.dots[i] = thisDot;
        }
    }
}
