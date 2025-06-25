module Data.Note exposing (CreateResponse, Note, decode, decodeCreateResponse)

import Json.Decode as D exposing (Decoder)


type alias CreateResponse =
    { slug : String }


decodeCreateResponse : Decoder CreateResponse
decodeCreateResponse =
    D.map CreateResponse
        (D.field "slug" D.string)


type alias Note =
    { content : String
    , readAt : Maybe String -- TODO: use Posix
    , createdAt : String -- TODO: use Posix
    , expiresAt : Maybe String -- TODO: use Posix
    }


decode : Decoder Note
decode =
    D.map4 Note
        (D.field "content" D.string)
        (D.field "read_at" (D.maybe D.string))
        (D.field "created_at" D.string)
        (D.field "expires_at" (D.maybe D.string))
