export interface AddTrackResponse {
    error?: string
    data?: Track
}

export interface GetTrackResponse {
    error?: string
    data: Track[]
}
export interface Track {
    id: string
    title: string
    artist_id: string
    file: string
    created_at: string
}