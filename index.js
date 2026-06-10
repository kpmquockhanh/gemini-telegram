const express = require("express");
const axios = require("axios");
require("dotenv").config();
const WorkerPool = require("./WorkerPool");

const pool = new WorkerPool(parseInt(process.env.WORKER_POOL_SIZE) || 4);

const api = axios.create({
  timeout: 300000,
  family: 4,
});

const app = express();
app.use(express.json());

const BOT_TOKEN = process.env.BOT_TOKEN;
const PORT = process.env.PORT || 3000;

const activeJobs = new Map();

app.get("/", (req, res) => {
  res.send("hello world");
});

async function sendMessage(chatId, text) {
  const res = await api.post(
    `https://api.telegram.org/bot${BOT_TOKEN}/sendMessage`,
    {
      chat_id: chatId,
      text,
    },
  );
  return res.data?.result?.message_id;
}

async function editMessage(chatId, messageId, text) {
  await api.post(`https://api.telegram.org/bot${BOT_TOKEN}/editMessageText`, {
    chat_id: chatId,
    message_id: messageId,
    text,
  });
}

async function sendPhoto(chatId, imageBuffer, caption) {
  const formData = new FormData();
  formData.append("chat_id", chatId);
  formData.append(
    "photo",
    new Blob([imageBuffer], { type: "image/png" }),
    "image.png",
  );
  if (caption) formData.append("caption", caption);

  await api.post(
    `https://api.telegram.org/bot${BOT_TOKEN}/sendPhoto`,
    formData,
    {
      headers: { "Content-Type": "multipart/form-data" },
    },
  );
}

async function sendVideo(chatId, videoBuffer) {
  const formData = new FormData();
  formData.append("chat_id", chatId);
  formData.append(
    "video",
    new Blob([videoBuffer], { type: "video/mp4" }),
    "video.mp4",
  );

  await api.post(
    `https://api.telegram.org/bot${BOT_TOKEN}/sendVideo`,
    formData,
    {
      headers: { "Content-Type": "multipart/form-data" },
    },
  );
}

async function downloadTelegramFile(fileId) {
  const { data: fileInfo } = await api.get(
    `https://api.telegram.org/bot${BOT_TOKEN}/getFile?file_id=${fileId}`,
  );
  const fileUrl = `https://api.telegram.org/file/bot${BOT_TOKEN}/${fileInfo.result.file_path}`;
  const response = await api.get(fileUrl, { responseType: "arraybuffer" });
  return Buffer.from(response.data);
}

app.post("/telegram", async (req, res) => {
  res.sendStatus(200);

  const message = req.body?.message;
  if (!message) return;

  const chatId = message.chat.id;
  const text = message.text || "";
  const caption = message.caption || "";

  if (text.startsWith("/image") || caption.startsWith("/image")) {
    const prompt = text.startsWith("/image")
      ? text.slice("/image".length).trim()
      : caption.slice("/image".length).trim();

    if (!prompt) {
      await sendMessage(
        chatId,
        "Usage: /image <prompt>\nOr reply to a photo with /image <prompt>",
      );
      return;
    }

    try {
      const messageId = await sendMessage(chatId, "🖼️ Generating image...");
      activeJobs.set(chatId, { messageId, startTime: Date.now() });

      let imageBuffer = null;
      let mimeType = "image/jpeg";

      if (message.reply_to_message?.photo) {
        const photo =
          message.reply_to_message.photo[
            message.reply_to_message.photo.length - 1
          ];
        imageBuffer = await downloadTelegramFile(photo.file_id);
        mimeType = "image/jpeg";
      } else if (message.photo) {
        const photo = message.photo[message.photo.length - 1];
        imageBuffer = await downloadTelegramFile(photo.file_id);
        mimeType = "image/jpeg";
      }

      const result = await pool.enqueue({
        type: "image",
        prompt,
        imageBuffer,
        mimeType,
        onProgress: async (elapsed) => {
          try {
            const job = activeJobs.get(chatId);
            if (job) {
              await editMessage(
                chatId,
                job.messageId,
                `🖼️ Generating image... (${elapsed}s)`,
              );
            }
          } catch {}
        },
      });

      activeJobs.delete(chatId);
      await sendPhoto(chatId, result, prompt);
    } catch (error) {
      activeJobs.delete(chatId);
      console.error("Image generation error:", error);
      await sendMessage(
        chatId,
        "Failed to generate image. Please try again later.",
      );
    }
    return;
  }

  if (text.startsWith("/video") || caption.startsWith("/video")) {
    const prompt = text.startsWith("/video")
      ? text.slice("/video".length).trim()
      : caption.slice("/video".length).trim();
    if (!prompt) {
      await sendMessage(
        chatId,
        "Usage: /video <prompt>\nOr reply to a photo with /video <prompt>",
      );
      return;
    }

    try {
      const messageId = await sendMessage(
        chatId,
        "🎬 Generating video... This may take a minute.",
      );
      activeJobs.set(chatId, { messageId, startTime: Date.now() });

      let imageBuffer = null;
      let mimeType = "image/png";

      if (message.reply_to_message?.photo) {
        const photo =
          message.reply_to_message.photo[
            message.reply_to_message.photo.length - 1
          ];
        imageBuffer = await downloadTelegramFile(photo.file_id);
        mimeType = "image/jpeg";
      } else if (message.photo) {
        const photo = message.photo[message.photo.length - 1];
        imageBuffer = await downloadTelegramFile(photo.file_id);
        mimeType = "image/jpeg";
      }

      const result = await pool.enqueue({
        type: "video",
        prompt,
        imageBuffer,
        mimeType,
        onProgress: async (elapsed) => {
          try {
            const job = activeJobs.get(chatId);
            if (job) {
              await editMessage(
                chatId,
                job.messageId,
                `🎬 Generating video... (${elapsed}s)`,
              );
            }
          } catch {}
        },
      });

      activeJobs.delete(chatId);
      await sendVideo(chatId, result);
    } catch (error) {
      activeJobs.delete(chatId);
      console.error(
        "Video generation error:",
        error.response?.data || error.message,
      );
      await sendMessage(
        chatId,
        "Failed to generate video. Please try again later.",
      );
    }
    return;
  }
});

process.on("SIGTERM", async () => {
  console.log("Shutting down worker pool...");
  await pool.shutdown();
  process.exit(0);
});

app.listen(PORT, () => {
  console.log(`Bot listening on port ${PORT}`);
});
