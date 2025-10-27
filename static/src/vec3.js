/**
 * Represents a 3D vector with x, y, z components.
 */
class Vec3 {
    /**
     * Creates a new Vec3.
     * @param {number} x - The x component.
     * @param {number} y - The y component.
     * @param {number} z - The z component.
     */
    constructor(x, y, z) {
        /** @type {number} */
        this.x = x;
        /** @type {number} */
        this.y = y;
        /** @type {number} */
        this.z = z;
    }

    /**
     * Adds another vector to this one and returns the result.
     * @param {Vec3} v - The vector to add.
     * @returns {Vec3} A new vector that is the sum.
     */
    add(v) {
        return new Vec3(this.x + v.x, this.y + v.y, this.z + v.z);
    }

    /**
     * Subtracts another vector from this one and returns the result.
     * @param {Vec3} v - The vector to subtract.
     * @returns {Vec3} A new vector that is the difference.
     */
    sub(v) {
        return new Vec3(this.x - v.x, this.y - v.y, this.z - v.z);
    }

    /**
     * Multiplies the vector by a scalar and returns the result.
     * @param {number} s - The scalar value.
     * @returns {Vec3} A new scaled vector.
     */
    mul(s) {
        return new Vec3(this.x * s, this.y * s, this.z * s);
    }

    /**
     * Divides the vector by a scalar and returns the result.
     * @param {number} s - The scalar value.
     * @returns {Vec3} A new scaled vector.
     */
    div(s) {
        return new Vec3(this.x / s, this.y / s, this.z / s);
    }

    /**
     * Returns the magnitude (length) of the vector.
     * @returns {number} The magnitude of the vector.
     */
    mag() {
        return Math.sqrt(this.x * this.x + this.y * this.y + this.z * this.z);
    }

    /**
     * Converts the vector into an `rgb(...)` string.
     * @return {string}
     */
    toRGB() {
        return `rgb(${this.x * 255}, ${this.y * 255}, ${this.z * 255})`
    }

    /**
     * Converts the vector into an `rgba(...)` string.
     * @param {number} opacity - The fourth component.
     * @return {string}
     */
    toRGBA(opacity) {
        return `rgba(${this.x * 255}, ${this.y * 255}, ${this.z * 255}, ${opacity})`
    }
}

/**
 * The zero vector (0, 0, 0). Can be used as the black color as well.
 * @type {Vec3}
 */
Vec3.zero = new Vec3(0, 0, 0);

/**
 * The unit vector with all components 1. Can be used as the white color as well.
 * @type {Vec3}
 */
Vec3.unit = new Vec3(1, 1, 1);

/**
 * Bulma primary dark color.
 * @type {Vec3}
 */
Vec3.bulmaPrimaryDark = new Vec3(0, 102 / 255, 87 / 255);

/**
 * Bulma primary color.
 * @type {Vec3}
 */
Vec3.bulmaPrimary = new Vec3(0, 209 / 255, 178 / 255);

/**
 * Bulma danger color.
 * @type {Vec3}
 */
Vec3.bulmaDanger = new Vec3(255 / 255, 56 / 255, 96 / 255);
