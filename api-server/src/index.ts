import express from "express"
import cors from "cors"
import { PORT } from "./lib/config"
import { RedisManager } from "./lib/redisManager"
import expressWs from "express-ws"
import { clientType, pubsubResType, type pubsubRes, type SSEResponse } from "./lib/types"
import { useSSEMiddleware } from "./lib/sseMiddleware"

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
        const cache_response = await instance.searchFromCache(search_query)
        if (cache_response) {
            const parsed_cache = JSON.parse(cache_response)

            res.json({ "query_id": parsed_cache.query_id, "response": parsed_cache.response })
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



app.get("/search/:queryId", useSSEMiddleware, async (req, res) => {
    const sse_res = res as SSEResponse

    try {
        //check the cache again to check if the worker has alr finished processing the request and added it to the cache
        const query_id = req.params.queryId as string
        console.log("query_id", query_id)
        if (!query_id.trim()) {
            sse_res.sendEventStreamData("error", { message: "Query id is required" });
            res.end();
            return
        }

        const normal_instance = await RedisManager.getInstance(clientType.normal)
        const search_query = await normal_instance.searchFromCache(query_id)
        if (search_query) {
            const cache_response = await normal_instance.searchFromCache(search_query)
            if (cache_response) {
                const parsed_cache = JSON.parse(cache_response)
                sse_res.sendEventStreamData("full_response", parsed_cache.response)
                res.end()
                return;
            }
        }

        const pubsub_instance = await RedisManager.getInstance(clientType.pubsub)
        //subscribe to the pubsub, waiting for the message
        await pubsub_instance.subscribe(query_id, async (data) => {
            const { type, payload } = JSON.parse(data) as pubsubRes

            sse_res.sendEventStreamData(type, payload)
            if (type === pubsubResType.full_response) {
                res.end()
                await pubsub_instance.unsubscribe(query_id)
            }
        })

        req.on('close', () => {
            res.end()
            pubsub_instance.unsubscribe(query_id)
        })
        //if the type is full response return full_response if the type is chunk stream the response
    } catch (error) {
        console.log("herererere...")
        const error_message = error instanceof Error ? error.message : "Something went wrong"
        sse_res.sendEventStreamData("error", { message: error_message });
        res.end();
    }
})


app.get("/health", (req, res) => {
    res.json({ message: "Server is healthy!" })
})

app.listen(PORT, () => {
    console.log(`Server is listening on port ${PORT}`)
})