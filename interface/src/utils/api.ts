import {BASE_URL} from "./constants.ts";

type RequestParams = {
    path: string
    method: HTTPMethods
    body?: Record<string, unknown>
}

enum HTTPMethods {
    get = "GET",
    post = "POST"
}

async function request<T>({ path, method = HTTPMethods.get, body} : RequestParams): Promise<T> {
    const response = await fetch(`${BASE_URL}${path}`, {
        method: method,
        headers: {
            'Content-Type': 'application/json',
        },
        body: body ? JSON.stringify(body) : undefined,
    })
    if(!response.ok){
        throw "Error while fetching"
    }
    return (await response.json()) as T
}