module Layouts.Header exposing (Model, Msg, Props, layout)

import Auth.User
import Effect exposing (Effect)
import Html exposing (Html)
import Html.Attributes as Attr
import Html.Events
import Layout exposing (Layout)
import Route exposing (Route)
import Route.Path
import Shared
import View exposing (View)


type alias Props =
    {}


layout : Props -> Shared.Model -> Route () -> Layout () Model Msg contentMsg
layout _ shared _ =
    Layout.new
        { init = init
        , update = update
        , view = view shared
        , subscriptions = subscriptions
        }



-- MODEL


type alias Model =
    {}


init : () -> ( Model, Effect Msg )
init _ =
    ( {}, Effect.none )



-- UPDATE


type Msg
    = UserClickedLogout


update : Msg -> Model -> ( Model, Effect Msg )
update msg model =
    case msg of
        UserClickedLogout ->
            ( model, Effect.logout )


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.none



-- VIEW


view : Shared.Model -> { toContentMsg : Msg -> contentMsg, content : View contentMsg, model : Model } -> View contentMsg
view shared { toContentMsg, content } =
    { title = content.title
    , body =
        [ viewNavbar shared |> Html.map toContentMsg
        , Html.main_ [] content.body
        ]
    }


viewNavbar : Shared.Model -> Html Msg
viewNavbar shared =
    Html.header [ Attr.class "navbar" ]
        [ Html.nav [ Attr.class "f-row justify-content:space-between" ]
            [ Html.ul [ Attr.attribute "role" "list" ]
                [ Html.li [] [ viewNavLink ( "home", Route.Path.Home_ ) ] ]
            , Html.ul [ Attr.attribute "role" "list" ]
                (case shared.user of
                    Auth.User.SignedIn _ ->
                        [ Html.li [] [ viewNavLink ( "profile", Route.Path.Profile_Me ) ]
                        , Html.li [] [ Html.a [ Html.Events.onClick UserClickedLogout ] [ Html.text "logout" ] ]
                        ]

                    Auth.User.NotSignedIn ->
                        viewNotSignedInNav

                    Auth.User.RefreshingTokens ->
                        viewNotSignedInNav
                )
            ]
        ]


viewNotSignedInNav : List (Html msg)
viewNotSignedInNav =
    [ Html.li [] [ viewNavLink ( "sign in", Route.Path.Auth ) ]
    ]


viewNavLink : ( String, Route.Path.Path ) -> Html msg
viewNavLink ( label, path ) =
    Html.a
        [ Route.Path.href path ]
        [ Html.text label ]
