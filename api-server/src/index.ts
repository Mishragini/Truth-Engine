import express from "express"
import cors from "cors"
import { PORT } from "./lib/config"
import { RedisManager } from "./lib/redisManager"
import expressWs from "express-ws"
import { clientType } from "./lib/types"

const { app } = expressWs(express())

app.use(express.json())
app.use(cors())

app.get("/search", async (req, res) => {
    try {
        const search_query = req.query.search_query as string
        if (!search_query) {
            res.status(400).json({ message: "search_query query param is required." })
            return
        }
        const instance = await RedisManager.getInstance(clientType.normal)
        const cache_response = await instance.searchFromCache(search_query as string)
        if (cache_response) {
            res.json({ "response": cache_response })
            return;
        }

        const queue_data = {
            search_query,
        }
        const query_id = await instance.enqueue(queue_data)
        res.json({ "query_id": query_id })

    } catch (error) {
        const error_message = error instanceof Error ? error.message : "Something went wrong"
        res.status(500).json({ "error_message": error_message })
    }
})

app.ws("/:requestId", async (ws, req) => {
    const redis_instance = await RedisManager.getInstance(clientType.pubsub)
    const request_id = req.params.requestId as string
    if (!request_id) {
        const message = {
            "success": false,
            "data": {
                "error": "Request id is required."
            }
        }
        ws.send(JSON.stringify(message))
    }

    ws.on('message', async (data: string) => {
        try {
            const { search_query } = JSON.parse(data)
            if (!search_query) {
                const message = {
                    "success": false,
                    "data": {
                        "error": "Search query is required."
                    }
                }
                ws.send(JSON.stringify(message))
            }
            const cache_response = await redis_instance.searchFromCache(search_query)
            if (cache_response) {
                const message = {
                    "success": true,
                    "data": {
                        "response": cache_response
                    }
                }
                ws.send(JSON.stringify(message))
            }

            await redis_instance.subscribe(request_id, (message) => {
                const { type, payload } = JSON.parse(message)
                const ws_message = {
                    "success": true,
                    "data": {
                        "type": type,
                        "payload": payload
                    }
                }
                ws.send(JSON.stringify(ws_message))
            })
        } catch (error) {
            const data = {
                "success": false,
                "data": {
                    "error": error instanceof Error ? error.message : "Something went wrong!"
                }
            }
            ws.send(JSON.stringify(data))
        }

    })


    ws.on('close', async () => {
        await redis_instance.unsubscribe(request_id)
    })

    ws.on('error', async () => {
        await redis_instance.unsubscribe(request_id)
    })

})

app.get("/health", (req, res) => {
    res.json({ message: "Server is healthy!" })
})

app.listen(PORT, () => {
    console.log(`Server is listening on port ${PORT}`)
})