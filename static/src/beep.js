class Beeper {
    constructor() {
        /**
         * @private
         * @type {AudioContext}
         */
        this.audioContext = new (window.AudioContext || window.webkitAudioContext)();

        /**
         * @private
         * @type {{ collision: AudioBuffer | null, death: AudioBuffer | null }}
         */
        this.buffers = {};

        this.muted = true;
    }

    /**
     * Toggle mute state, and resume the context if suspended.
     * @return {Promise<boolean>} - True if audio is now muted, false otherwise.
     */
    async toggleMute() {
        // Resume if necessary (browsers block until interaction).
        if (this.audioContext.state === "suspended") await this.audioContext.resume();

        // Toggle mute status.
        this.muted = !this.muted;
        return this.muted;
    }

    /**
     * Preload a collision sound into a buffer. This sound can be played using the "playCollision" method.
     * This can be called even if the AudioContext is suspended.
     */
    async loadCollisionSound() {
        const duration = 0.1;
        const sampleRate = this.audioContext.sampleRate;
        const buffer = this.audioContext.createBuffer(1, sampleRate * duration, sampleRate);
        const data = buffer.getChannelData(0);

        for (let i = 0; i < data.length; i++) {
            const t = i / sampleRate;

            // Frequency ramp from 800Hz to 200Hz over duration.
            const startFreq = 800;
            const endFreq = 200;
            const frequency = startFreq * Math.pow(endFreq / startFreq, t / duration);

            // Gain envelope: starts at 0.3, exponentially decays to 0.01
            const gain = 0.3 * Math.pow(0.01 / 0.3, t / duration);
            // Generate the waveform.
            data[i] = Math.sin(2 * Math.PI * frequency * t) * gain;
        }

        this.buffers.collision = buffer;
    }

    /**
     * Preload a death sound into a buffer. This sound can be played using the "playDeath" method.
     * This can be called even if the AudioContext is suspended.
     */
    async loadDeathSound() {
        const duration = 0.1;
        const sampleRate = this.audioContext.sampleRate;
        const buffer = this.audioContext.createBuffer(1, sampleRate * duration, sampleRate);
        const data = buffer.getChannelData(0);

        for (let i = 0; i < data.length; i++) {
            const t = i / sampleRate;

            // Frequency ramp from 400Hz to 50Hz over duration (lower, more ominous).
            const startFreq = 400;
            const endFreq = 50;
            const frequency = startFreq * Math.pow(endFreq / startFreq, t / duration);

            // Gain envelope: starts at 0.2, exponentially decays to 0.001
            const gain = 0.2 * Math.pow(0.001 / 0.2, t / duration);
            // Generate the waveform.
            data[i] = Math.sin(2 * Math.PI * frequency * t) * gain;
        }

        this.buffers.death = buffer;
    }

    /**
     * Play the collision sound.
     */
    playCollision() {
        // Respect the mute setting.
        if (this.muted) return;

        // Collision sound needs to be loaded beforehand using the "loadCollisionSound" method.
        if (!this.buffers.collision) {
            console.warn("collision audio buffer is not loaded");
            return;
        }

        if (this.audioContext.state === "suspended") {
            console.warn("AudioContext state is suspended. Some user interaction is required to enable it");
            return;
        }

        const source = this.audioContext.createBufferSource();
        source.buffer = this.buffers.collision;
        source.connect(this.audioContext.destination);
        source.start();
    }

    /**
     * Play the death sound.
     */
    playDeath() {
        // Respect the mute setting.
        if (this.muted) return;

        // Death sound needs to be loaded beforehand using the "loadDeathSound" method.
        if (!this.buffers.death) {
            console.warn("death audio buffer is not loaded");
            return;
        }

        if (this.audioContext.state === "suspended") {
            console.warn("AudioContext state is suspended. Some user interaction is required to enable it");
            return;
        }

        const source = this.audioContext.createBufferSource();
        source.buffer = this.buffers.death;
        source.connect(this.audioContext.destination);
        source.start();
    }
}
