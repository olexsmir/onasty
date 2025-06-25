module Pages.Secret.Slug_ exposing (Model, Msg, page)

import Api
import Api.Note
import Components.Error
import Components.Note
import Data.Note exposing (Metadata, Note)
import Effect exposing (Effect)
import Html as H exposing (Html)
import Html.Attributes as A
import Html.Events as E
import Layouts
import Page exposing (Page)
import Route exposing (Route)
import Shared
import View exposing (View)


page : Shared.Model -> Route { slug : String } -> Page Model Msg
page _ route =
    Page.new
        { init = init route.params.slug
        , update = update
        , subscriptions = subscriptions
        , view = view
        }
        |> Page.withLayout (\_ -> Layouts.Header {})



-- INIT


type alias Model =
    { slug : String
    , note : Maybe (Api.Response Note)
    , metadata : Api.Response Metadata
    }


init : String -> () -> ( Model, Effect Msg )
init slug () =
    ( { slug = slug
      , note = Nothing
      , metadata = Api.Loading
      }
    , Api.Note.fetchMetadata
        { onResponse = ApiGetMetadataResponded
        , slug = slug
        }
    )



-- UPDATE


type Msg
    = UserClickedViewNote
    | ApiGetNoteResponded (Result Api.Error Note)
    | ApiGetMetadataResponded (Result Api.Error Metadata)


update : Msg -> Model -> ( Model, Effect Msg )
update msg model =
    case msg of
        UserClickedViewNote ->
            ( { model | note = Just Api.Loading }
            , Api.Note.get
                { onResponse = ApiGetNoteResponded
                , slug = model.slug
                }
            )

        ApiGetNoteResponded (Ok note) ->
            ( { model | note = Just (Api.Success note) }, Effect.none )

        ApiGetNoteResponded (Err error) ->
            ( { model | note = Just (Api.Failure error) }, Effect.none )

        ApiGetMetadataResponded (Ok metadata) ->
            ( { model | metadata = Api.Success metadata }, Effect.none )

        ApiGetMetadataResponded (Err error) ->
            ( { model | metadata = Api.Failure error }, Effect.none )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.none



-- VIEW


type PageVariant
    = RequestNote
    | ShowNote


view : Model -> View Msg
view model =
    { title = "View note"
    , body =
        [ H.div [ A.class "py-8 px-4" ]
            [ H.div [ A.class "w-full max-w-4xl mx-auto" ]
                [ H.div [ A.class "bg-white rounded-lg border border-gray-200 shadow-sm" ]
                    (case model.note of
                        Nothing ->
                            [ viewHeader GetNote
                            , viewOpenNote False
                            ]

                        Just Api.Loading ->
                            [ viewHeader GetNote
                            , viewOpenNote True
                            ]

                        Just (Api.Success note) ->
                            [ viewHeader ViewNote
                            , viewNoteContent note
                            ]

                        Just (Api.Failure err) ->
                            [ Components.Error.error (Api.errorMessage err) ]
                    )
                ]
            ]
        ]
    }



-- HEADER


type HeaderVariant
    = GetNote
    | ViewNote


viewHeader : HeaderVariant -> Html msg
viewHeader variant =
    H.div [ A.class "p-6 pb-4 border-b border-gray-200" ]
        (case variant of
            GetNote ->
                [ H.h1
                    [ A.class "text-2xl font-bold text-gray-900" ]
                    [ H.text "View note" ]
                , H.p [ A.class "text-gray-600 mt-2" ]
                    [ H.text "Click the button below to view the note content" ]
                ]

            ViewNote ->
                [ H.div [ A.class "flex justify-between items-start" ]
                    [ H.div []
                        [ H.h1 [ A.class "text-2xl font-bold text-gray-900" ] [ H.text "Note: Slug" ]
                        , H.div [ A.class "text-sm text-gray-500 mt-2 space-y-1" ]
                            [ H.p [] [ H.text "Created at: 2023-10-01T12:00:00Z" ]
                            , H.p [] [ H.text "Expires at: 2023-10-01T12:00:00Z" ]
                            ]
                        ]
                    , H.div [ A.class "flex gap-2" ]
                        [ H.button [ A.class "px-3 py-2 text-sm border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2 transition-colors" ]
                            [ H.text "Copy Content" ]
                        ]
                    ]
                ]
        )



-- NOTE


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
    H.div [ A.class "p-6" ]
        [ H.div [ A.class "bg-gray-50 border border-gray-200 rounded-md p-4" ]
            [ H.pre [ A.class "whitespace-pre-wrap font-mono text-sm text-gray-800 overflow-x-auto" ] [ H.text note.content ] ]
        ]
