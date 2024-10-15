module Api exposing (Response)


type Response
    = NoteCreated String
    | NoteCreateFailed String
    | NoteRead String
    | NoteNotFound
