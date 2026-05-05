import { type RouteConfig, index, route } from "@react-router/dev/routes";

export default [index("routes/home.tsx"), route("/chat/:query_id", "routes/chat/$query_id.tsx")] satisfies RouteConfig;
