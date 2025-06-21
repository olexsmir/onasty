module Effect exposing
    ( Effect
    , none, batch
    , sendCmd, sendMsg
    , pushRoute, replaceRoute
    , pushRoutePath, replaceRoutePath
    , loadExternalUrl, back
    , sendApiRequest, refreshTokens
    , signin, logout, saveUser, clearUser
    , map, toCmd
    )

{-|

@docs Effect

@docs none, batch
@docs sendCmd, sendMsg

@docs pushRoute, replaceRoute
@docs pushRoutePath, replaceRoutePath
@docs loadExternalUrl, back

@docs sendApiRequest, refreshTokens
@docs signin, logout, saveUser, clearUser

@docs map, toCmd

-}

import Api
import Auth.User
import Browser.Navigation
import Data.Credentials exposing (Credentials)
import Data.Error
import Dict exposing (Dict)
import Http
import Json.Decode
import Json.Encode
import Ports exposing (sendToLocalStorage)
import Route
import Route.Path
import Shared.Model
import Shared.Msg
import Task
import Url exposing (Url)


type Effect msg
    = -- BASICS
      None
    | Batch (List (Effect msg))
    | SendCmd (Cmd msg)
      -- ROUTING
    | PushUrl String
    | ReplaceUrl String
    | LoadExternalUrl String
    | Back
      -- SHARED
    | SendSharedMsg Shared.Msg.Msg
    | SendToLocalStorage { key : String, value : Json.Encode.Value }
    | SendApiRequest (HttpRequestDetails msg)


type alias HttpRequestDetails msg =
    { endpoint : String
    , method : String
    , body : Http.Body
    , decoder : Json.Decode.Decoder msg
    , onHttpError : Api.Error -> msg
    }



-- BASICS


{-| Don't send any effect.
-}
none : Effect msg
none =
    None


{-| Send multiple effects at once.
-}
batch : List (Effect msg) -> Effect msg
batch =
    Batch


{-| Send a normal `Cmd msg` as an effect, something like `Http.get` or `Random.generate`.
-}
sendCmd : Cmd msg -> Effect msg
sendCmd =
    SendCmd


{-| Send a message as an effect. Useful when emitting events from UI components.
-}
sendMsg : msg -> Effect msg
sendMsg msg =
    Task.succeed msg
        |> Task.perform identity
        |> SendCmd



-- ROUTING


{-| Set the new route, and make the back button go back to the current route.
-}
pushRoute :
    { path : Route.Path.Path
    , query : Dict String String
    , hash : Maybe String
    }
    -> Effect msg
pushRoute route =
    PushUrl (Route.toString route)


{-| Same as `Effect.pushRoute`, but without `query` or `hash` support
-}
pushRoutePath : Route.Path.Path -> Effect msg
pushRoutePath path =
    PushUrl (Route.Path.toString path)


{-| Set the new route, but replace the previous one, so clicking the back
button **won't** go back to the previous route.
-}
replaceRoute :
    { path : Route.Path.Path
    , query : Dict String String
    , hash : Maybe String
    }
    -> Effect msg
replaceRoute route =
    ReplaceUrl (Route.toString route)


{-| Same as `Effect.replaceRoute`, but without `query` or `hash` support
-}
replaceRoutePath : Route.Path.Path -> Effect msg
replaceRoutePath path =
    ReplaceUrl (Route.Path.toString path)


{-| Redirect users to a new URL, somewhere external to your web application.
-}
loadExternalUrl : String -> Effect msg
loadExternalUrl =
    LoadExternalUrl


{-| Navigate back one page
-}
back : Effect msg
back =
    Back



-- SHARED


sendApiRequest :
    { endpoint : String
    , method : String
    , body : Http.Body
    , decoder : Json.Decode.Decoder value
    , onResponse : Result Api.Error value -> msg
    }
    -> Effect msg
sendApiRequest opts =
    let
        onHttpError : Api.Error -> msg
        onHttpError err =
            opts.onResponse (Err err)

        decoder : Json.Decode.Decoder msg
        decoder =
            opts.decoder
                |> Json.Decode.map Ok
                |> Json.Decode.map opts.onResponse
    in
    SendApiRequest
        { endpoint = opts.endpoint
        , method = opts.method
        , body = opts.body
        , onHttpError = onHttpError
        , decoder = decoder
        }


refreshTokens : Effect msg
refreshTokens =
    SendSharedMsg Shared.Msg.TriggerTokenRefresh


signin : Credentials -> Effect msg
signin credentials =
    SendSharedMsg (Shared.Msg.SignedIn credentials)


logout : Effect msg
logout =
    SendSharedMsg Shared.Msg.Logout


saveUser : String -> String -> Effect msg
saveUser accessToken refreshToken =
    batch
        [ SendToLocalStorage { key = "access_token", value = Json.Encode.string accessToken }
        , SendToLocalStorage { key = "refresh_token", value = Json.Encode.string refreshToken }
        ]


clearUser : Effect msg
clearUser =
    batch
        [ SendToLocalStorage { key = "access_token", value = Json.Encode.null }
        , SendToLocalStorage { key = "refresh_token", value = Json.Encode.null }
        ]



-- INTERNALS


{-| Elm Land depends on this function to connect pages and layouts
together into the overall app.
-}
map : (msg1 -> msg2) -> Effect msg1 -> Effect msg2
map fn effect =
    case effect of
        None ->
            None

        Batch list ->
            Batch (List.map (map fn) list)

        SendCmd cmd ->
            SendCmd (Cmd.map fn cmd)

        PushUrl url ->
            PushUrl url

        ReplaceUrl url ->
            ReplaceUrl url

        Back ->
            Back

        LoadExternalUrl url ->
            LoadExternalUrl url

        SendSharedMsg sharedMsg ->
            SendSharedMsg sharedMsg

        SendToLocalStorage options ->
            SendToLocalStorage options

        SendApiRequest opts ->
            SendApiRequest
                { endpoint = opts.endpoint
                , method = opts.method
                , body = opts.body
                , decoder = Json.Decode.map fn opts.decoder
                , onHttpError = \err -> fn (opts.onHttpError err)
                }


{-| Elm Land depends on this function to perform your effects.
-}
toCmd :
    { key : Browser.Navigation.Key
    , url : Url
    , shared : Shared.Model.Model
    , fromSharedMsg : Shared.Msg.Msg -> msg
    , batch : List msg -> msg
    , toCmd : msg -> Cmd msg
    }
    -> Effect msg
    -> Cmd msg
toCmd options effect =
    case effect of
        None ->
            Cmd.none

        Batch list ->
            Cmd.batch (List.map (toCmd options) list)

        SendCmd cmd ->
            cmd

        PushUrl url ->
            Browser.Navigation.pushUrl options.key url

        ReplaceUrl url ->
            Browser.Navigation.replaceUrl options.key url

        Back ->
            Browser.Navigation.back options.key 1

        LoadExternalUrl url ->
            Browser.Navigation.load url

        SendSharedMsg sharedMsg ->
            Task.succeed sharedMsg
                |> Task.perform options.fromSharedMsg

        SendToLocalStorage opts ->
            sendToLocalStorage opts

        SendApiRequest opts ->
            let
                headers : List Http.Header
                headers =
                    case options.shared.user of
                        Auth.User.SignedIn cred ->
                            if not (String.contains opts.endpoint "refresh-tokens") then
                                [ Http.header "Authorization" ("Bearer " ++ cred.accessToken) ]

                            else
                                []

                        _ ->
                            []
            in
            Http.request
                { method = opts.method
                , url = opts.endpoint
                , headers = headers
                , body = opts.body
                , expect =
                    Http.expectStringResponse
                        (\httpResult ->
                            case httpResult of
                                Ok msg ->
                                    msg

                                Err err ->
                                    opts.onHttpError err
                        )
                        (\resp -> fromHttpResponseToCustomError opts.decoder resp)
                , timeout = Just (1000 * 60) -- 60 second timeout
                , tracker = Nothing
                }


fromHttpResponseToCustomError : Json.Decode.Decoder msg -> Http.Response String -> Result Api.Error msg
fromHttpResponseToCustomError decoder response =
    case response of
        Http.GoodStatus_ _ body ->
            case
                Json.Decode.decodeString decoder
                    (if String.isEmpty body then
                        "\"\""

                     else
                        body
                    )
            of
                Ok value ->
                    Ok value

                Err err ->
                    Err (Api.JsonDecodeError { message = "Failed to decode JSON response", reason = err })

        Http.BadStatus_ { statusCode } body ->
            case Json.Decode.decodeString Data.Error.decode body of
                Ok err ->
                    Err (Api.HttpError { message = err.message, reason = Http.BadStatus statusCode })

                Err err ->
                    Err (Api.JsonDecodeError { message = "Something unexpected happened", reason = err })

        Http.BadUrl_ url ->
            Err (Api.HttpError { message = "Unexpected URL format", reason = Http.BadUrl url })

        Http.Timeout_ ->
            Err (Api.HttpError { message = "Request timed out, please try again", reason = Http.Timeout })

        Http.NetworkError_ ->
            Err (Api.HttpError { message = "Could not connect, please try again", reason = Http.NetworkError })
