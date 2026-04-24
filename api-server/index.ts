import express from "express"
import cors from "cors"
import { PORT } from "./lib/config"
import { RedisManager } from "./lib/redisManager"
const app = express()

app.use(express.json())
app.use(cors())

app.get("/search", async (req, res) => {
    try {
        const { search_query } = req.query
        if (!search_query) {
            res.status(400).json({ message: "search_query query param is required." })
            return
        }
        const instance = await RedisManager.getInstance()
        const cache_response = await instance.searchFromCache(search_query as string)
        if (cache_response) {
            res.json({ "response": cache_response })
            return;
        }
        const query_id = await instance.enqueue(search_query as string)
        res.json({ "query_id": query_id })
    } catch (error) {
        const error_message = error instanceof Error ? error.message : "Something went wrong"
        res.status(500).json({ "error_message": error_message })
    }
})


app.get("/health", (req, res) => {
    res.json({ message: "Server is healthy!" })
})

app.listen(PORT, () => {
    console.log(`Server is listening on port ${PORT}`)
})