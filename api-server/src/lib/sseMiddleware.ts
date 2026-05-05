import type { Request, Response, NextFunction } from "express";
import { FRONTEND_URL } from "./config";
import type { SSEResponse } from "./types";

export const useSSEMiddleware = (req: Request, res: Response, next: NextFunction) => {
    res.setHeader('Content-Type', "text/event-stream");
    res.setHeader('Cache-Control', 'no-cache');
    //Some proxies(nginx, etc.) will close the connection after a short idle.
    res.setHeader('Connection', 'keep-alive');

    res.setHeader('Access-Control-Allow-Origin', FRONTEND_URL)

    res.flushHeaders();

    const sendEventStreamData = (event: string, data: unknown) => {
        const sseFormattedResponse = `event: ${event}\ndata: ${JSON.stringify(data)}\n\n`
        res.write(sseFormattedResponse)
    }

    (res as SSEResponse).sendEventStreamData = sendEventStreamData

    next();
}