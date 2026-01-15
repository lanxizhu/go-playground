import axios from "axios";
import { MD5 } from "crypto-js";

// --- DOM Elements ---
const dropzone = document.getElementById("dropzone") as HTMLDivElement;
const fileInput = document.getElementById("file") as HTMLInputElement;
const fileInfo = document.getElementById("file-info") as HTMLDivElement;
const fileName = document.getElementById("file-name") as HTMLDivElement;
const fileSize = document.getElementById("file-size") as HTMLDivElement;
const progressContainer = document.getElementById("progress-container") as HTMLDivElement;
const progressBar = document.getElementById("progress-bar") as HTMLDivElement;
const progressText = document.getElementById("progress-text") as HTMLSpanElement;
const progressPercent = document.getElementById("progress-percent") as HTMLSpanElement;
const messageDiv = document.getElementById("message") as HTMLDivElement;
const uploadForm = document.getElementById("upload-form") as HTMLFormElement;
const resetBtn = document.getElementById("reset-btn") as HTMLButtonElement;

// --- Constants ---
const CHUNK_SIZE = 1024 * 1024 * 5; // 5MB
const BATCH_SIZE = 5;

// --- Toast Logic ---
const createToastContainer = () => {
  let container = document.getElementById("toast-container");
  if (!container) {
    container = document.createElement("div");
    container.id = "toast-container";
    container.className = "toast-container";
    document.body.appendChild(container);
  }
  return container;
};

const showToast = (message: string, type: "success" | "error" | "info" = "info") => {
  const container = createToastContainer();
  const toast = document.createElement("div");
  toast.className = `toast ${type}`;

  const iconMap = {
    success: "✅",
    error: "❌",
    info: "ℹ️"
  };

  toast.innerHTML = `
    <span class="toast-icon">${iconMap[type]}</span>
    <span class="toast-message">${message}</span>
  `;

  container.appendChild(toast);

  // Trigger animation
  requestAnimationFrame(() => {
    toast.classList.add("show");
  });

  // Remove after 3 seconds
  setTimeout(() => {
    toast.classList.remove("show");
    setTimeout(() => {
      container.removeChild(toast);
    }, 300);
  }, 3000);
};

// --- Helper Functions ---
const formatSize = (bytes: number): string => {
  if (bytes === 0) return "0 Bytes";
  const k = 1024;
  const sizes = ["Bytes", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
};

const updateProgress = (uploadedChunks: number, totalChunks: number) => {
  const percent = Math.round((uploadedChunks / totalChunks) * 100);
  progressBar.style.width = `${percent}%`;
  progressPercent.innerText = `${percent}%`;
  progressText.innerText = percent === 100 ? "Processing..." : "Uploading...";
};

const resetUI = () => {
  fileInput.value = "";
  fileInfo.classList.remove("active");
  progressContainer.classList.remove("active");
  progressBar.style.width = "0%";
  fileName.innerText = "";
  fileSize.innerText = "";
  dropzone.classList.remove("drag-active");
};

// --- Event Listeners ---

// Dropzone Drag & Drop
dropzone.addEventListener("click", () => fileInput.click());

dropzone.addEventListener("dragover", (e) => {
  e.preventDefault();
  dropzone.classList.add("drag-active");
});

dropzone.addEventListener("dragleave", () => {
  dropzone.classList.remove("drag-active");
});

dropzone.addEventListener("drop", (e) => {
  e.preventDefault();
  dropzone.classList.remove("drag-active");
  if (e.dataTransfer && e.dataTransfer.files.length > 0) {
    fileInput.files = e.dataTransfer.files;
    handleFileSelect();
  }
});

// File Selection
fileInput.addEventListener("change", handleFileSelect);

function handleFileSelect() {
  const file = fileInput.files?.[0];
  if (file) {
    fileInfo.classList.add("active");
    fileName.innerText = file.name;
    fileSize.innerText = formatSize(file.size);
  }
}

// Reset
resetBtn.addEventListener("click", (e) => {
  e.preventDefault();
  resetUI();
});

// Upload Logic
uploadForm.addEventListener("submit", async (event) => {
  event.preventDefault();

  const file = fileInput.files?.[0];

  // TOAST CHECK
  if (!file) {
    showToast("Please select a file first!", "error");
    return;
  }

  // Initial UI Setup for Upload
  progressContainer.classList.add("active");
  let chunksCompleted = 0;

  const chunks = createChunks(file);
  const totalChunks = chunks.length;

  try {
    // 1. Check Status
    const md5Hash = MD5(file.name).toString(); // Note: Ideally hash content, but using name for speed as per original logic
    const statusRes = await axios.get("upload/status", {
      headers: { "X-File-Id": md5Hash }
    });

    const uploadedSet = new Set(statusRes.data.uploaded);
    chunksCompleted = uploadedSet.size;
    updateProgress(chunksCompleted, totalChunks);

    if (chunksCompleted === totalChunks) {
      await completeUpload(file, md5Hash, totalChunks);
      return;
    }

    // 2. Prepare Tasks
    const tasks = chunks.map((chunk, index) => {
      const chunkNum = index + 1;
      if (uploadedSet.has(chunkNum)) {
        return () => Promise.resolve();
      }

      const formData = new FormData();
      formData.append("file", chunk.file);

      return () =>
        axios.post("upload/check", formData, {
          headers: {
            "Content-Type": "multipart/form-data",
            "X-File-Id": md5Hash,
            "X-Chunk-Number": chunkNum,
            "X-Total-Chunks": totalChunks,
          }
        }).then(() => {
          chunksCompleted++;
          updateProgress(chunksCompleted, totalChunks);
        });
    });

    // 3. Execute in Batches
    await sendRequestsInBatches(tasks);

    // 4. Complete
    await completeUpload(file, md5Hash, totalChunks);

  } catch (error) {
    console.error("Upload failed", error);
    showToast("Upload failed. Please try again.", "error");
  }
});

const createChunks = (file: File, size = CHUNK_SIZE) => {
  const chunkList: { file: Blob }[] = [];
  let cur = 0;
  while (cur < file.size) {
    chunkList.push({ file: file.slice(cur, cur + size) });
    cur += size;
  }
  return chunkList;
};

async function sendRequestsInBatches(tasks: Array<() => Promise<void>>, batchSize = BATCH_SIZE) {
  for (let i = 0; i < tasks.length; i += batchSize) {
    const batch = tasks.slice(i, i + batchSize);
    await Promise.all(batch.map(task => task()));
  }
}

async function completeUpload(file: File, fileId: string, totalChunks: number) {
  try {
    const res = await axios.post("upload/complete", null, {
      headers: {
        "X-File-Id": fileId,
        "X-File-Name": encodeURIComponent(file.name),
        "X-Total-Chunks": totalChunks,
        "Content-Type": "multipart/form-data",
      }
    });
    console.log("Upload Complete:", res.data);
    updateProgress(totalChunks, totalChunks); // Ensure 100%
    showToast("File uploaded successfully!", "success");
    setTimeout(() => {
      // Optional: Reset after success
      // resetUI();
    }, 3000);
  } catch (error) {
    console.error("Completion failed", error);
    showToast("File assembly failed.", "error");
  }
}
