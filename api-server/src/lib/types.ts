import type { Response } from "express"

export enum clientType {
    normal = "normal",
    pubsub = "pubsub"
}

export enum pubsubResType {
    full_response = "full_response",
    chunk = "chunk"
}
interface pubsubRes {
    type: pubsubResType,
    payload: unknown
}
interface SSEResponse extends Response {
    sendEventStreamData: (event: string, data: unknown) => void
}

export type { pubsubRes, SSEResponse }