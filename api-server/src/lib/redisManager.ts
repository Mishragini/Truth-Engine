import { createClient, type RedisClientType } from "redis";
import { REDIS_URL } from "./config";
import { clientType } from "./types";

export class RedisManager {
    private static instance: RedisManager;
    private normalClient: RedisClientType
    private pubsubClient: RedisClientType

    private constructor() {
        this.normalClient = createClient({
            url: REDIS_URL
        })

        this.pubsubClient = createClient({
            url: REDIS_URL
        })
    }

    private async connectQueue() {
        if (!this.normalClient.isOpen) {
            await this.normalClient.connect()
        }
    }

    private async connectPubSub() {
        if (!this.pubsubClient.isOpen) {
            await this.pubsubClient.connect()
        }
    }

    public static async getInstance(client: clientType) {
        if (!RedisManager.instance) {
            RedisManager.instance = new RedisManager()
        }
        client === clientType.normal
            ? await this.instance.connectQueue()
            : await this.instance.connectPubSub()
        return RedisManager.instance
    }

    public async searchFromCache(cache_key: string) {
        const cache_search_response = await this.normalClient.get(cache_key)

        if (cache_search_response) {
            return cache_search_response
        }
        return null
    }

    public async enqueue(data: { search_query: string }) {
        const query_id = `job_${Date.now()}_${Math.random().toString(36).substring(2, 15)}`;
        const query_data = {
            query_id,
            search_query: data.search_query,
        }
        const pipeline = this.normalClient.multi()
        pipeline.lPush("search_query", JSON.stringify(query_data))
        pipeline.set(query_id, data.search_query, { EX: 24 * 3600 })
        await pipeline.exec()
        return query_id
    }

    public async subscribe(task_id: string, callback: (message: string) => void) {
        await this.pubsubClient.subscribe(task_id, callback)
    }

    public async unsubscribe(task_id: string) {
        await this.pubsubClient.unsubscribe(task_id)
    }
}