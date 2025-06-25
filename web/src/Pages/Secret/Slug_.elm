module Pages.Secret.Slug_ exposing (Model, Msg, page)

import Api
import Components.Error
import Components.Note
import Data.Note exposing (Note)
import Effect exposing (Effect)
import Html as H exposing (Html, button)
import Html.Attributes as A
import Html.Events as E
import Layouts
import Page exposing (Page)
import Route exposing (Route)
import Shared
import View exposing (View)


page : Shared.Model -> Route { slug : String } -> Page Model Msg
page shared route =
    Page.new
        { init = init
        , update = update
        , subscriptions = subscriptions
        , view = view
        }
        |> Page.withLayout (\_ -> Layouts.Header {})



-- INIT


type alias Model =
    { note : Maybe (Api.Response Note) }


init : () -> ( Model, Effect Msg )
init () =
    ( { note = Nothing }
    , Effect.none
    )



-- UPDATE


type Msg
    = UserClickedViewNote


update : Msg -> Model -> ( Model, Effect Msg )
update msg model =
    case msg of
        UserClickedViewNote ->
            ( { model | note = Just Api.Loading }, Effect.none )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.none



-- VIEW


view : Model -> View Msg
view model =
    { title = "View note"
    , body =
        [ H.div [ A.class "py-8 px-4" ]
            [ H.div [ A.class "w-full max-w-4xl mx-auto" ]
                [ H.div [ A.class "bg-white rounded-lg border border-gray-200 shadow-sm" ]
                    [ viewHeader
                    , case model.note of
                        Nothing ->
                            viewOpenNote False

                        Just Api.Loading ->
                            viewOpenNote True

                        Just (Api.Success note) ->
                            viewNoteContent note

                        Just (Api.Failure err) ->
                            Components.Error.error (Api.errorMessage err)
                    ]
                ]
            ]
        ]
    }


viewHeader : Html msg
viewHeader =
    H.div [ A.class "p-6 pb-4 border-b border-gray-200" ]
        [ H.h1
            [ A.class "text-2xl font-bold text-gray-900" ]
            [ H.text "View note" ]
        , H.p [ A.class "text-gray-600 mt-2" ]
            [ H.text "Click the button below to view the note content" ]
        ]


viewOpenNote : Bool -> Html Msg
viewOpenNote isLoading =
    let
        buttonData : { text : String, class : String }
        buttonData =
            let
                base : String
                base =
                    "px-6 py-3 rounded-md focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2 transition-colors"
            in
            if isLoading then
                { text = "Loading Note...", class = base ++ " bg-gray-300 text-gray-500 cursor-not-allowed" }

            else
                { text = "View Note", class = base ++ " bg-black text-white hover:bg-gray-800" }
    in
    H.div [ A.class "p-6" ]
        [ H.div [ A.class "text-center py-12" ]
            [ H.div [ A.class "mb-6" ]
                [ H.div [ A.class "w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4" ]
                    [ Components.Note.noteIconSvg ]
                , H.h2 [ A.class "text-lg font-semibold text-gray-900 mb-2" ] [ H.text "note slug" ]
                , H.p [ A.class "text-gray-600 mb-6" ] [ H.text "This note is protected. Click below to view its content." ]
                ]
            , H.button
                [ A.class buttonData.class
                , E.onClick UserClickedViewNote
                ]
                [ H.text buttonData.text ]
            ]
        ]


viewNoteContent : Note -> Html msg
viewNoteContent note =
    H.div [] []
