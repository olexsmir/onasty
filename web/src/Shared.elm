module Shared exposing
    ( Flags, decoder
    , Model, Msg
    , init, update, subscriptions
    )

{-|

@docs Flags, decoder
@docs Model, Msg
@docs init, update, subscriptions

-}

import Api.Auth
import Auth.User
import Data.Credentials exposing (Credentials)
import Dict
import Effect exposing (Effect)
import Json.Decode
import JwtUtil
import Route exposing (Route)
import Route.Path
import Shared.Model
import Shared.Msg
import Task
import Time



-- FLAGS


type alias Flags =
    { accessToken : Maybe String
    , refreshToken : Maybe String
    }


decoder : Json.Decode.Decoder Flags
decoder =
    Json.Decode.map2 Flags
        (Json.Decode.field "access_token" (Json.Decode.maybe Json.Decode.string))
        (Json.Decode.field "refresh_token" (Json.Decode.maybe Json.Decode.string))



-- INIT


type alias Model =
    Shared.Model.Model


init : Result Json.Decode.Error Flags -> Route () -> ( Model, Effect Msg )
init flagsResult _ =
    let
        flags : Flags
        flags =
            flagsResult |> Result.withDefault { accessToken = Nothing, refreshToken = Nothing }

        maybeCredentials : Maybe Credentials
        maybeCredentials =
            Maybe.map2
                (\access refresh -> { accessToken = access, refreshToken = refresh })
                flags.accessToken
                flags.refreshToken

        user : Auth.User.SignInStatus
        user =
            case maybeCredentials of
                Just credentials ->
                    Auth.User.SignedIn credentials

                Nothing ->
                    Auth.User.NotSignedIn

        initModel : Model
        initModel =
            { user = user
            , timeZone = Time.utc
            }
    in
    ( initModel
    , Effect.batch
        [ Time.now |> Task.perform Shared.Msg.CheckTokenExpiration |> Effect.sendCmd
        , Time.here |> Task.perform Shared.Msg.GotZone |> Effect.sendCmd
        ]
    )



-- UPDATE


type alias Msg =
    Shared.Msg.Msg


update : Route () -> Msg -> Model -> ( Model, Effect Msg )
update _ msg model =
    case msg of
        Shared.Msg.GotZone timeZone ->
            ( { model | timeZone = timeZone }, Effect.none )

        Shared.Msg.Logout ->
            ( { model | user = Auth.User.NotSignedIn }, Effect.clearUser )

        Shared.Msg.SignedIn credentials ->
            ( { model | user = Auth.User.SignedIn credentials }
            , Effect.batch
                [ Effect.pushRoute
                    { path = Route.Path.Home_
                    , query = Dict.empty
                    , hash = Nothing
                    }
                , Effect.saveUser credentials.accessToken credentials.refreshToken
                ]
            )

        Shared.Msg.CheckTokenExpiration now ->
            case model.user of
                Auth.User.SignedIn credentials ->
                    if JwtUtil.isExpired now credentials.accessToken then
                        ( model, Effect.refreshTokens )

                    else
                        ( model, Effect.none )

                _ ->
                    ( model, Effect.none )

        Shared.Msg.TriggerTokenRefresh ->
            case model.user of
                Auth.User.SignedIn credentials ->
                    ( { model | user = Auth.User.RefreshingTokens }
                    , Api.Auth.refreshToken
                        { onResponse = Shared.Msg.ApiRefreshTokensResponded
                        , refreshToken = credentials.refreshToken
                        }
                    )

                _ ->
                    ( model, Effect.none )

        Shared.Msg.ApiRefreshTokensResponded (Ok credentials) ->
            ( { model | user = Auth.User.SignedIn credentials }
            , Effect.saveUser credentials.accessToken credentials.refreshToken
            )

        Shared.Msg.ApiRefreshTokensResponded (Err _) ->
            ( model, Effect.clearUser )



-- SUBSCRIPTIONS


subscriptions : Route () -> Model -> Sub Msg
subscriptions _ _ =
    Time.every (30 * 1000) Shared.Msg.CheckTokenExpiration
