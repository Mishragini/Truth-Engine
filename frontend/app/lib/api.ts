import axios from "axios"

export async function performSearch(search_query: string) {
    console.log("here....", import.meta.env.VITE_BACKEND_BASE_URL)
    let url = import.meta.env.VITE_BACKEND_BASE_URL + `/search?search_query=${search_query}`
    let api_response = await axios.get(url)
    console.log("api_response...", api_response)
    if (api_response.status !== 200) {
        let error_message = api_response.data.error_message || "Failed to perform search."
        throw new Error(error_message)
    }
    return api_response
}