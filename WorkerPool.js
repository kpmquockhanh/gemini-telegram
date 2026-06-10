const { Worker } = require("worker_threads");
const path = require("node:path");

const WORKER_PATH = path.join(__dirname, "workers", "ai-worker.js");

class WorkerPool {
  constructor(size = 4) {
    this.size = size;
    this.workers = [];
    this.queue = [];
    this.active = new Map();
    this.jobCounter = 0;
    this.init();
  }

  init() {
    for (let i = 0; i < this.size; i++) {
      this.spawnWorker();
    }
  }

  spawnWorker() {
    const worker = new Worker(WORKER_PATH);
    this.workers.push(worker);

    worker.on("message", (msg) => {
      const job = this.active.get(worker);
      if (!job) return;

      if (msg.type === "progress") {
        job.onProgress?.(msg.elapsed);
      } else if (msg.type === "done") {
        this.active.delete(worker);
        job.resolve(msg.result);
        this.processQueue(worker);
      } else if (msg.type === "error") {
        this.active.delete(worker);
        job.reject(new Error(msg.error));
        this.processQueue(worker);
      }
    });

    worker.on("error", (err) => {
      const job = this.active.get(worker);
      if (job) {
        this.active.delete(worker);
        job.reject(err);
      }
      this.replaceWorker(worker);
    });

    worker.on("exit", () => {
      const idx = this.workers.indexOf(worker);
      if (idx !== -1) {
        this.workers.splice(idx, 1);
        this.replaceWorker(worker);
      }
    });
  }

  replaceWorker(deadWorker) {
    const idx = this.workers.indexOf(deadWorker);
    if (idx !== -1) this.workers.splice(idx, 1);
    this.spawnWorker();
    if (this.queue.length > 0) {
      this.processQueue(this.workers[this.workers.length - 1]);
    }
  }

  processQueue(worker) {
    if (this.queue.length === 0) return;
    const next = this.queue.shift();
    this.dispatch(worker, next);
  }

  dispatch(worker, job) {
    this.active.set(worker, job);
    const payload = {
      jobId: job.jobId,
      type: job.type,
      prompt: job.prompt,
      imageBuffer: job.imageBuffer,
      mimeType: job.mimeType,
    };
    if (job.imageBuffer) {
      worker.postMessage(payload, [job.imageBuffer.buffer]);
    } else {
      worker.postMessage(payload);
    }
  }

  getAvailableWorker() {
    return this.workers.find((w) => !this.active.has(w));
  }

  enqueue({ type, prompt, imageBuffer, mimeType, onProgress }) {
    return new Promise((resolve, reject) => {
      const jobId = ++this.jobCounter;
      const job = { jobId, type, prompt, imageBuffer, mimeType, resolve, reject, onProgress };

      const worker = this.getAvailableWorker();
      if (worker) {
        this.dispatch(worker, job);
      } else {
        this.queue.push(job);
      }
    });
  }

  async shutdown() {
    for (const worker of this.workers) {
      worker.terminate();
    }
    this.workers = [];
    this.queue = [];
    this.active.clear();
  }
}

module.exports = WorkerPool;
