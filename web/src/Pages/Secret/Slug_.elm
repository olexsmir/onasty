module Pages.Secret.Slug_ exposing (Model, Msg, PageVariant, page)

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


type PageVariant
    = RequestNote
    | ShowNote
    | NotFound


type alias Model =
    { slug : String
    , page : PageVariant
    , note : Maybe (Api.Response Note)
    , metadata : Api.Response Metadata
    }


init : String -> () -> ( Model, Effect Msg )
init slug () =
    ( { slug = slug
      , page = RequestNote
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
    | UserClickedCopyContent
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

        UserClickedCopyContent ->
            ( model, Effect.none )

        ApiGetNoteResponded (Ok note) ->
            ( { model | page = ShowNote, note = Just (Api.Success note) }, Effect.none )

        ApiGetNoteResponded (Err error) ->
            ( { model | note = Just (Api.Failure error) }, Effect.none )

        ApiGetMetadataResponded (Ok metadata) ->
            ( { model | metadata = Api.Success metadata }, Effect.none )

        ApiGetMetadataResponded (Err error) ->
            ( { model | page = NotFound, metadata = Api.Failure error }, Effect.none )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.none



-- VIEW


view : Model -> View Msg
view model =
    { title = "View note"
    , body =
        [ H.div [ A.class "py-8 px-4" ]
            [ H.div [ A.class "w-full max-w-4xl mx-auto" ]
                [ H.div [ A.class "bg-white rounded-lg border border-gray-200 shadow-sm" ]
                    (let
                        notFound : List (Html msg)
                        notFound =
                            [ viewHeader { title = "Note Not Found", subtitle = "The note you're looking for doesn't exist or has expired" }
                            , viewNoteNotFound model.slug
                            ]
                     in
                     case model.page of
                        RequestNote ->
                            [ viewHeader { title = "View note", subtitle = "Click the button below to view the note content" }
                            , viewOpenNote False
                            ]

                        ShowNote ->
                            let
                                generic : List (Html Msg)
                                generic =
                                    [ viewHeader { title = "View note", subtitle = "Click the button below to view the note content" }
                                    , viewOpenNote True
                                    ]
                            in
                            case model.note of
                                Nothing ->
                                    generic

                                Just Api.Loading ->
                                    generic

                                Just (Api.Failure err) ->
                                    if Api.is404 err then
                                        notFound

                                    else
                                        [ Components.Error.error (Api.errorMessage err) ]

                                Just (Api.Success note) ->
                                    [ viewShowNoteHeader model.slug note
                                    , viewNoteContent note
                                    ]

                        NotFound ->
                            notFound
                    )
                ]
            ]
        ]
    }



-- HEADER


viewHeader : { title : String, subtitle : String } -> Html msg
viewHeader options =
    H.div [ A.class "p-6 pb-4 border-b border-gray-200" ]
        [ H.h1
            [ A.class "text-2xl font-bold text-gray-900" ]
            [ H.text options.title ]
        , H.p [ A.class "text-gray-600 mt-2" ] [ H.text options.subtitle ]
        ]


viewShowNoteHeader : String -> Note -> Html Msg
viewShowNoteHeader slug note =
    H.div [ A.class "p-6 pb-4 border-b border-gray-200" ]
        [ H.div [ A.class "flex justify-between items-start" ]
            [ H.div []
                [ H.h1 [ A.class "text-2xl font-bold text-gray-900" ] [ H.text ("Note: " ++ slug) ]
                , H.div [ A.class "text-sm text-gray-500 mt-2 space-y-1" ]
                    [ H.p [] [ H.text ("Created" ++ note.createdAt) ]
                    , case note.expiresAt of
                        Just expiresAt ->
                            H.p [] [ H.text ("Expires at: " ++ expiresAt) ]

                        Nothing ->
                            H.text ""
                    ]
                ]
            , H.div [ A.class "flex gap-2" ]
                [ H.button
                    [ E.onClick UserClickedCopyContent
                    , A.class "px-3 py-2 text-sm border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2 transition-colors"
                    ]
                    [ H.text "Copy Content" ]
                ]
            ]
        ]



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
            [ H.pre
                [ A.class "whitespace-pre-wrap font-mono text-sm text-gray-800 overflow-x-auto" ]
                [ H.text note.content ]
            ]
        ]


viewNoteNotFound : String -> Html msg
viewNoteNotFound slug =
    H.div [ A.class "p-6" ]
        [ H.div [ A.class "text-center py-12" ]
            [ H.div [ A.class "w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4" ]
                [ Components.Note.noteNotFoundSvg ]
            , H.h2 [ A.class "text-xl font-semibold text-gray-900 mb-2" ]
                [ H.text ("Note " ++ slug ++ " Not Found") ]
            , H.div [ A.class "text-gray-600 mb-6 space-y-2" ]
                [ H.p []
                    [ H.text "This note may have:"
                    , H.ul [ A.class "text-sm space-y-1 list-disc list-inside text-left max-w-md mx-auto" ]
                        [ H.li [] [ H.text "Expired and been automatically deleted" ]
                        , H.li [] [ H.text "Been deleted by the creator" ]
                        , H.li [] [ H.text "Been burned after reading (if it was a one-time view)" ]
                        , H.li [] [ H.text "Never existed or the URL is incorrect" ]
                        ]
                    ]
                ]
            ]
        ]
