import { createClient, type RedisClientType } from "redis";
import { REDIS_URL } from "./config";

export class RedisManager {
    private static instance: RedisManager;
    private client: RedisClientType

    private constructor() {
        this.client = createClient({
            url: REDIS_URL
        })
    }

    private async connectRedis() {
        if (!this.client.isOpen) {
            await this.client.connect()
        }
    }

    public static async getInstance() {
        if (!RedisManager.instance) {
            RedisManager.instance = new RedisManager()
            await this.instance.connectRedis()
        }
        return RedisManager.instance
    }

    public async searchFromCache(search_query: string) {
        const cache_key = search_query
        const cache_search_response = await this.client.get(cache_key)

        if (cache_search_response) {
            return JSON.parse(cache_search_response)
        }
        return null
    }

    public async enqueue(search_query: string) {
        const query_id = `job_${Date.now()}_${Math.random().toString(36).substring(2, 15)}`;
        const query_data = {
            query_id,
            search_query
        }
        await this.client.lPush("search_query", JSON.stringify(query_data))
        return query_id
    }
}