module Api exposing (..)

import Api.Endpoint as Endpoint exposing (Endpoint)
import Http
import Json.Decode exposing (Decoder)


{-| The authentication credentials for the Viewer (that is, the currently logged-in user.)
This simply includes the JWT token.
-}
type Cred
    = Cred String


credHeader : Cred -> Http.Header
credHeader (Cred c) =
    Http.header "Authorization" ("Bearer " ++ c)



-- http


get : Endpoint -> Maybe Cred -> Decoder a -> Http.Request a
get url cred decoder =
    Endpoint.request
        { method = "GET"
        , url = url
        , expect = Http.expectJson decoder
        , headers =
            case cred of
                Just c ->
                    [ credHeader c ]

                Nothing ->
                    []
        , body = Http.emptyBody
        , timeout = Nothing
        , withCredentials = False
        }


post : Endpoint -> Maybe Cred -> Http.Body -> Decoder a -> Http.Request a
post url cred body decoder =
    Endpoint.request
        { method = "POST"
        , url = url
        , expect = Http.expectJson decoder
        , headers =
            case cred of
                Just c ->
                    [ credHeader c ]

                Nothing ->
                    []
        , body = body
        , timeout = Nothing
        , withCredentials = False
        }
