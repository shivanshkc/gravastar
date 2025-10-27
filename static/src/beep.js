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

            // Sharp frequency drop from 600Hz to 80Hz (more dramatic).
            const startFreq = 600;
            const endFreq = 80;
            const frequency = startFreq * Math.pow(endFreq / startFreq, Math.pow(t / duration, 2));

            // Sharp attack then quick decay
            let gain;
            if (t < 0.02) {
                gain = 0.4 * (t / 0.02); // Quick attack
            } else {
                gain = 0.4 * Math.pow(0.001 / 0.4, (t - 0.02) / (duration - 0.02)); // Sharp decay
            }

            // Add some noise for a more organic death sound
            const noise = (Math.random() - 0.5) * 0.1 * gain;
            
            // Generate the waveform with some harmonics
            const fundamental = Math.sin(2 * Math.PI * frequency * t);
            const harmonic = Math.sin(2 * Math.PI * frequency * 2 * t) * 0.3;
            
            data[i] = (fundamental + harmonic) * gain + noise;
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
