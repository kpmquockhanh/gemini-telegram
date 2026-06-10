const { parentPort } = require("worker_threads");
const { GoogleGenAI } = require("@google/genai");
const fs = require("node:fs");
const path = require("node:path");
const os = require("node:os");

const ai = new GoogleGenAI({ apiKey: process.env.GEMINI_API_KEY });

const PROGRESS_INTERVAL = 10000;

function startProgressTimer(startTime) {
  const timer = setInterval(() => {
    const elapsed = Math.floor((Date.now() - startTime) / 1000);
    parentPort.postMessage({ type: "progress", elapsed });
  }, PROGRESS_INTERVAL);
  return timer;
}

async function generateImage(prompt, imageBuffer, mimeType) {
  const parts = [{ text: prompt }];

  if (imageBuffer) {
    const base64Image = Buffer.from(imageBuffer).toString("base64");
    parts.push({
      inlineData: {
        mimeType: mimeType || "image/jpeg",
        data: base64Image,
      },
    });
  }

  const response = await ai.models.generateContent({
    model: "gemini-3.1-flash-image",
    contents: parts,
  });

  for (const part of response.candidates[0].content.parts) {
    if (part.inlineData) {
      return Buffer.from(part.inlineData.data, "base64");
    }
  }

  return null;
}

async function generateVideo(prompt, imageBuffer, mimeType) {
  const params = {
    model: "veo-3.1-generate-preview",
    prompt: prompt,
    config: { durationSeconds: 4 },
  };

  if (imageBuffer) {
    params.image = {
      imageBytes: Buffer.from(imageBuffer).toString("base64"),
      mimeType: mimeType || "image/png",
    };
  }

  let operation = await ai.models.generateVideos(params);

  while (!operation.done) {
    await new Promise((resolve) => setTimeout(resolve, 10000));
    operation = await ai.operations.getVideosOperation({ operation });
  }

  if (operation.response?.generatedVideos?.[0]?.video) {
    const tmpPath = path.join(os.tmpdir(), `veo-${Date.now()}.mp4`);
    await ai.files.download({
      file: operation.response.generatedVideos[0].video,
      downloadPath: tmpPath,
    });
    const buffer = fs.readFileSync(tmpPath);
    fs.unlinkSync(tmpPath);
    return buffer;
  }

  return null;
}

parentPort.on("message", async ({ jobId, type, prompt, imageBuffer, mimeType }) => {
  const startTime = Date.now();
  const timer = startProgressTimer(startTime);

  try {
    let result = null;

    if (type === "image") {
      result = await generateImage(prompt, imageBuffer, mimeType);
    } else if (type === "video") {
      result = await generateVideo(prompt, imageBuffer, mimeType);
    }

    clearInterval(timer);

    if (result) {
      parentPort.postMessage({
        type: "done",
        jobId,
        result: result,
      }, [result.buffer]);
    } else {
      parentPort.postMessage({
        type: "error",
        jobId,
        error: `Failed to generate ${type}`,
      });
    }
  } catch (err) {
    clearInterval(timer);
    parentPort.postMessage({
      type: "error",
      jobId,
      error: err.message || "Unknown error",
    });
  }
});
