module Data.Note exposing (CreateResponse, Metadata, Note, decode, decodeCreateResponse, decodeMetadata)

import Iso8601
import Json.Decode as D exposing (Decoder)
import Time exposing (Posix)


type alias CreateResponse =
    { slug : String }


decodeCreateResponse : Decoder CreateResponse
decodeCreateResponse =
    D.map CreateResponse (D.field "slug" D.string)


type alias Note =
    { slug : String
    , content : String
    , readAt : Maybe Posix
    , keepBeforeExpiration : Bool
    , hasPassword : Bool
    , createdAt : Posix
    , expiresAt : Maybe Posix
    }


decode : Decoder Note
decode =
    D.map7 Note
        (D.field "slug" D.string)
        (D.field "content" D.string)
        (D.maybe (D.field "read_at" Iso8601.decoder))
        (D.field "keep_before_expiration" D.bool)
        (D.field "has_password" D.bool)
        (D.field "created_at" Iso8601.decoder)
        (D.maybe (D.field "expires_at" Iso8601.decoder))


type alias Metadata =
    { createdAt : Posix
    , hasPassword : Bool
    }


decodeMetadata : Decoder Metadata
decodeMetadata =
    D.map2 Metadata
        (D.field "created_at" Iso8601.decoder)
        (D.field "has_password" D.bool)
