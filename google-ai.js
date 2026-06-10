const { GoogleGenAI } = require("@google/genai");
const fs = require("node:fs");
const path = require("node:path");
const os = require("node:os");

const ai = new GoogleGenAI({
    apiKey: process.env.GEMINI_API_KEY,
});

async function generateImage(prompt, imageBuffer, mimeType) {
    const parts = [{ text: prompt }];

    if (imageBuffer) {
        const base64Image = imageBuffer.toString("base64");
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
        config: {
            durationSeconds: 4,
        }
    };

    if (imageBuffer) {
        params.image = {
            imageBytes: imageBuffer.toString("base64"),
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

module.exports = { generateImage, generateVideo };
