import dotenv from "dotenv"

dotenv.config()

export const PORT = process.env.PORT || 3000
export const REDIS_URL = process.env.REDIS_HOST || "redis://localhost:6379"