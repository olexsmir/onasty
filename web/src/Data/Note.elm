module Data.Note exposing (CreateResponse, Metadata, Note, decode, decodeCreateResponse, decodeMetadata)

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
    , burnBeforeExpiration : Maybe Bool
    , createdAt : String -- TODO: use Posix
    , expiresAt : Maybe String -- TODO: use Posix
    }


decode : Decoder Note
decode =
    D.map5 Note
        (D.field "content" D.string)
        (D.maybe (D.field "read_at" D.string))
        (D.maybe (D.field "burn_before_expiration" D.bool))
        (D.field "created_at" D.string)
        (D.maybe (D.field "expires_at" D.string))


type alias Metadata =
    { createdAt : String -- TODO: use Posix
    , hasPassword : Bool
    }


decodeMetadata : Decoder Metadata
decodeMetadata =
    D.map2 Metadata
        (D.field "created_at" D.string)
        (D.field "has_password" D.bool)
