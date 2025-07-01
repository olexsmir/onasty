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
    | ShowNote (Api.Response Note)
    | NotFound


type alias Model =
    { page : PageVariant
    , metadata : Api.Response Metadata
    , slug : String
    , password : Maybe String
    }


init : String -> () -> ( Model, Effect Msg )
init slug () =
    ( { page = RequestNote
      , metadata = Api.Loading
      , slug = slug
      , password = Nothing
      }
    , Api.Note.getMetadata
        { onResponse = ApiGetMetadataResponded
        , slug = slug
        }
    )



-- UPDATE


type Msg
    = UserClickedViewNote
    | UserClickedCopyContent
    | UserUpdatedPassword String
    | ApiGetNoteResponded (Result Api.Error Note)
    | ApiGetMetadataResponded (Result Api.Error Metadata)


update : Msg -> Model -> ( Model, Effect Msg )
update msg model =
    case msg of
        UserClickedViewNote ->
            ( { model | page = ShowNote Api.Loading }
            , Api.Note.get
                { onResponse = ApiGetNoteResponded
                , password = model.password
                , slug = model.slug
                }
            )

        UserClickedCopyContent ->
            case model.page of
                ShowNote (Api.Success note) ->
                    ( model, Effect.sendToClipboard note.content )

                _ ->
                    ( model, Effect.none )

        UserUpdatedPassword password ->
            ( { model | password = Just password }, Effect.none )

        ApiGetNoteResponded (Ok note) ->
            ( { model | page = ShowNote (Api.Success note) }, Effect.none )

        ApiGetNoteResponded (Err error) ->
            ( { model | page = ShowNote (Api.Failure error) }, Effect.none )

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
        [ H.div
            [ A.class "w-full max-w-4xl mx-auto" ]
            [ H.div
                [ A.class "bg-white rounded-lg border border-gray-200 shadow-sm" ]
                (case model.metadata of
                    Api.Success metadata ->
                        viewPage model.slug model.page metadata model.password

                    Api.Loading ->
                        [ viewHeader { title = "View note", subtitle = "Loading note metadata..." }
                        , viewOpenNote { slug = model.slug, hasPassword = False, password = Nothing, isLoading = True }
                        ]

                    Api.Failure error ->
                        [ viewHeader { title = "Note Not Found", subtitle = "The note you're looking for doesn't exist or has expired" }
                        , if Api.is404 error then
                            viewNoteNotFound model.slug

                          else
                            Components.Error.error (Api.errorMessage error)
                        ]
                )
            ]
        ]
    }


viewPage : String -> PageVariant -> Metadata -> Maybe String -> List (Html Msg)
viewPage slug variant metadata password =
    case variant of
        RequestNote ->
            [ viewHeader { title = "View note", subtitle = "Click the button below to view the note content" }
            , viewOpenNote { slug = slug, hasPassword = metadata.hasPassword, password = password, isLoading = False }
            ]

        ShowNote apiResp ->
            case apiResp of
                Api.Success note ->
                    [ viewShowNoteHeader slug note
                    , viewNoteContent note
                    ]

                Api.Loading ->
                    [ viewHeader { title = "View note", subtitle = "Click the button below to view the note content" }
                    , viewOpenNote { slug = slug, hasPassword = metadata.hasPassword, password = password, isLoading = True }
                    ]

                Api.Failure _ ->
                    [ viewHeader { title = "Note Not Found", subtitle = "The note you're looking for doesn't exist or has expired" }
                    , viewNoteNotFound slug
                    ]

        NotFound ->
            [ viewNoteNotFound slug ]



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
    H.div []
        [ if note.burnBeforeExpiration then
            H.div [ A.class "bg-orange-50 border-b border-orange-200 p-4" ]
                [ H.div [ A.class "flex items-center gap-3" ]
                    [ H.div [ A.class "w-6 h-6 bg-orange-100 rounded-full flex items-center justify-center flex-shrink-0" ]
                        [ Components.Note.warningSvg ]
                    , H.p [ A.class "text-orange-800 text-sm font-medium" ]
                        [ H.text "This note was destroyed. If you need to keep it, copy it before closing this window." ]
                    ]
                ]

          else
            H.text ""
        , H.div [ A.class "p-6 pb-4 border-b border-gray-200" ]
            [ H.div [ A.class "flex justify-between items-start" ]
                [ H.div []
                    [ H.h1 [ A.class "text-2xl font-bold text-gray-900" ] [ H.text ("Note: " ++ slug) ]
                    , H.div [ A.class "text-sm text-gray-500 mt-2 space-y-1" ]
                        [ H.p [] [ H.text ("Created: " ++ note.createdAt) ]
                        , case note.expiresAt of
                            Just expiresAt ->
                                -- TODO: format time properly
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
        ]



-- NOTE


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
                    [ H.span [ A.class "font-bold" ] [ H.text "This note may have:" ]
                    , H.ul [ A.class "text-sm space-y-1 list-disc list-inside text-left max-w-md mx-auto" ]
                        [ H.li [] [ H.text "Expired and been deleted" ]
                        , H.li [] [ H.text "Have different password" ]
                        , H.li [] [ H.text "Been deleted by the creator" ]
                        , H.li [] [ H.text "Been burned after reading" ]
                        , H.li [] [ H.text "Never existed or the URL is incorrect" ]
                        ]
                    ]
                ]
            ]
        ]


viewOpenNote :
    { slug : String
    , hasPassword : Bool
    , isLoading : Bool
    , password : Maybe String
    }
    -> Html Msg
viewOpenNote opts =
    let
        isDisabled =
            opts.hasPassword && Maybe.withDefault "" opts.password == ""

        buttonData =
            let
                base =
                    "px-6 py-3 rounded-md focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2 transition-colors"
            in
            if opts.isLoading then
                { text = "Loading Note...", class = base ++ " bg-gray-300 text-gray-500 cursor-not-allowed" }

            else if isDisabled then
                { text = "View Note", class = base ++ " bg-gray-300 text-gray-500 cursor-not-allowed" }

            else
                { text = "View Note", class = base ++ " bg-black text-white hover:bg-gray-800" }
    in
    H.div [ A.class "p-6" ]
        [ H.div [ A.class "text-center py-12" ]
            [ H.div [ A.class "mb-6" ]
                [ H.div [ A.class "w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4" ]
                    [ Components.Note.noteIconSvg ]
                , H.h2 [ A.class "text-lg font-semibold text-gray-900 mb-2" ] [ H.text opts.slug ]
                , H.p [ A.class "text-gray-600 mb-6" ] [ H.text "You're about read and destroy the note." ]
                ]
            , H.form
                [ E.onSubmit UserClickedViewNote
                , A.class "max-w-sm mx-auto space-y-4"
                ]
                [ if opts.hasPassword then
                    H.div
                        [ A.class "space-y-2" ]
                        [ H.label
                            [ A.class "block text-sm font-medium text-gray-700 text-left" ]
                            [ H.text "Password" ]
                        , H.input
                            [ E.onInput UserUpdatedPassword
                            , A.class "w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-black focus:border-transparent"
                            ]
                            []
                        ]

                  else
                    H.text ""
                , H.button
                    [ A.class buttonData.class
                    , A.type_ "submit"
                    , A.disabled isDisabled
                    ]
                    [ H.text buttonData.text ]
                ]
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
