module Layouts.Header exposing (Model, Msg, Props, layout)

import Auth.User
import Effect exposing (Effect)
import Html as H exposing (Html)
import Html.Attributes as A
import Html.Events as E
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
        [ viewHeader shared.user |> H.map toContentMsg
        , H.main_ [] content.body
        ]
    }


viewHeader : Auth.User.SignInStatus -> Html Msg
viewHeader user =
    H.header [ A.class "w-full border-b border-gray-200 bg-white" ]
        [ H.div [ A.class "max-w-7xl mx-auto px-4 sm:px-6 lg:px-8" ]
            [ H.div [ A.class "flex justify-between items-center h-16" ]
                [ H.div [ A.class "flex items-center" ]
                    [ H.a
                        [ A.class "text-lg font-semibold text-black hover:text-gray-700 transition-colors"
                        , Route.Path.href Route.Path.Home_
                        ]
                        [ H.text "Onasty" ]
                    ]
                , H.nav [ A.class "flex items-center space-x-6" ] (viewNav user)
                ]
            ]
        ]


viewNav : Auth.User.SignInStatus -> List (Html Msg)
viewNav user =
    let
        viewLink text path =
            H.a [ A.class "text-gray-600 hover:text-black transition-colors", Route.Path.href path ]
                [ H.text text ]
    in
    case user of
        Auth.User.SignedIn _ ->
            [ viewLink "Profile" Route.Path.Profile_Me
            , H.button
                [ A.class "text-gray-600 hover:text-red-600 transition-colors"
                , E.onClick UserClickedLogout
                ]
                [ H.text "Logout" ]
            ]

        _ ->
            [ viewLink "About" Route.Path.Home_ -- TODO: or add about page, or delete the link
            , H.a
                [ A.class "px-4 py-2 border border-gray-300 rounded-md text-black hover:bg-gray-50 transition-colors"
                , Route.Path.href Route.Path.Auth
                ]
                [ H.text "Sign In/Up" ]
            ]
