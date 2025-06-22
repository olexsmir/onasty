module Pages.Home_ exposing (Model, Msg, page)

import Effect exposing (Effect)
import Html as H exposing (Html)
import Html.Attributes as A
import Html.Events as E
import Layouts
import Page exposing (Page)
import Route exposing (Route)
import Shared
import View exposing (View)


page : Shared.Model -> Route () -> Page Model Msg
page shared _ =
    Page.new
        { init = init shared
        , update = update
        , subscriptions = subscriptions
        , view = view shared
        }
        |> Page.withLayout Layouts.Header



-- INIT


type alias Model =
    {}


init : Shared.Model -> () -> ( Model, Effect Msg )
init _ () =
    ( {}, Effect.none )



-- UPDATE


type Msg
    = NoOp


update : Msg -> Model -> ( Model, Effect Msg )
update msg model =
    case msg of
        NoOp ->
            ( model, Effect.none )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.none



-- VIEW


view : Shared.Model -> Model -> View Msg
view _ _ =
    { title = "Onasty"
    , body =
        [ H.div [ A.class "py-8 px-4 " ]
            [ H.div [ A.class "w-full max-w-4xl mx-auto" ]
                [ H.div [ A.class "bg-white rounded-lg border border-gray-200 shadow-sm" ]
                    [ viewHeader
                    , H.div [ A.class "p-6 space-y-6" ]
                        [ viewForm ]
                    ]
                ]
            ]
        ]
    }


viewHeader : Html Msg
viewHeader =
    H.div [ A.class "p-6 pb-4 border-b border-gray-200" ]
        [ H.h1 [ A.class "text-2xl font-bold text-gray-900" ]
            [ H.text "Create new note" ]
        ]


viewForm : Html Msg
viewForm =
    -- TODO: that form defo should be broken down into smaller components
    H.form
        [ E.onSubmit NoOp -- TODO: implement me
        , A.class "space-y-6"
        ]
        [ H.div [ A.class "space-y-2" ]
            [ H.label
                [ A.for "content", A.class "block text-sm font-medium text-gray-700 mb-2" ]
                [ H.text "Content *" ]
            , H.textarea
                [ A.id "content"
                , A.class "w-full h-96 px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-black focus:border-transparent resize-vertical font-mono text-sm"
                , A.placeholder "Write your note here..."
                , A.required True
                ]
                []
            ]
        , H.div [ A.class "space-y-2" ]
            [ H.label [ A.for "slug", A.class "block text-sm font-medium text-gray-700 mb-2" ] [ H.text "Custom URL Slug (optional)" ]
            , H.input
                [ A.id "slug"
                , A.type_ "text"
                , A.placeholder "my-unique-slug"
                , A.class "w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-black focus:border-transparent"
                ]
                []
            , H.p [ A.class "text-xs text-gray-500 mt-1" ] [ H.text "Leave empty to generate a random slug" ]
            ]
        , H.div
            [ A.class "flex justify-end" ]
            [ viewSubmitButton ]
        ]


viewSubmitButton : Html Msg
viewSubmitButton =
    H.button
        [ A.type_ "submit"
        , A.disabled True -- TODO: check if form is valid to be sent
        , A.class
            (if True then
                "px-6 py-2 bg-gray-300 text-gray-500 rounded-md cursor-not-allowed transition-colors"

             else
                "px-6 py-2 bg-black text-white rounded-md hover:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2 transition-colors"
            )
        ]
        [ H.text "Create note" ]
