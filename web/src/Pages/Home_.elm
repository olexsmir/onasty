module Pages.Home_ exposing (Model, Msg, page, PageVariant)

import Api
import Api.Note
import Data.Note as Note
import Effect exposing (Effect)
import Html as H exposing (Html)
import Html.Attributes as A
import Html.Events as E
import Layouts
import Page exposing (Page)
import Process
import Route exposing (Route)
import Shared
import Task
import View exposing (View)


page : Shared.Model -> Route () -> Page Model Msg
page shared _ =
    Page.new
        { init = init shared
        , update = update shared
        , subscriptions = subscriptions
        , view = view shared
        }
        |> Page.withLayout (\_ -> Layouts.Header {})



-- INIT


type alias Model =
    { pageVariant : PageVariant
    , content : String
    , slug : Maybe String
    , apiError : Maybe Api.Error
    , userClickedCopyLink : Bool
    }


type PageVariant
    = CreateNote
    | NoteCreated String


init : Shared.Model -> () -> ( Model, Effect Msg )
init _ () =
    ( { pageVariant = CreateNote
      , content = ""
      , slug = Nothing
      , userClickedCopyLink = False
      , apiError = Nothing
      }
    , Effect.none
    )



-- UPDATE


type Msg
    = UserUpdatedInput Field String
    | UserClickedSubmit
    | UserClickedCreateNewNote
    | UserClickedCopyLink
    | CopiedFeedbackShown
    | ApiCreateNoteResponded (Result Api.Error Note.CreateResponse)


type Field
    = Content
    | Slug


update : Shared.Model -> Msg -> Model -> ( Model, Effect Msg )
update shared msg model =
    case msg of
        UserClickedSubmit ->
            ( model
            , Api.Note.create
                { onResponse = ApiCreateNoteResponded
                , content = model.content
                , slug = model.slug
                }
            )

        UserClickedCreateNewNote ->
            ( { model
                | pageVariant = CreateNote
                , content = ""
                , slug = Nothing
                , apiError = Nothing
              }
            , Effect.none
            )

        UserClickedCopyLink ->
            ( { model | userClickedCopyLink = True }
            , Effect.batch
                [ Effect.sendCmd (Task.perform (\_ -> CopiedFeedbackShown) (Process.sleep 2000))
                , Effect.sendToClipboard (secretUrl shared.appURL (Maybe.withDefault "" model.slug))
                ]
            )

        CopiedFeedbackShown ->
            ( { model | userClickedCopyLink = False }, Effect.none )

        UserUpdatedInput Content content ->
            ( { model | content = content }, Effect.none )

        UserUpdatedInput Slug slug ->
            ( { model | slug = Just slug }, Effect.none )

        ApiCreateNoteResponded (Ok response) ->
            ( { model | pageVariant = NoteCreated response.slug, slug = Just response.slug }, Effect.none )

        ApiCreateNoteResponded (Err error) ->
            ( { model | apiError = Just error }, Effect.none )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.none



-- VIEW
-- TODO: show errors
-- TODO: validate form


view : Shared.Model -> Model -> View Msg
view shared model =
    { title = "Onasty"
    , body =
        [ H.div [ A.class "py-8 px-4 " ]
            [ H.div [ A.class "w-full max-w-4xl mx-auto" ]
                [ H.div [ A.class "bg-white rounded-lg border border-gray-200 shadow-sm" ]
                    [ viewHeader
                    , H.div [ A.class "p-6 space-y-6" ]
                        (case model.pageVariant of
                            CreateNote ->
                                [ viewForm model ]

                            NoteCreated slug ->
                                [ viewNoteCreated model.userClickedCopyLink shared.appURL slug ]
                        )
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


viewForm : Model -> Html Msg
viewForm model =
    -- TODO: that form defo should be broken down into smaller components
    H.form
        [ E.onSubmit UserClickedSubmit
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
                , E.onInput (UserUpdatedInput Content)
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
                , E.onInput (UserUpdatedInput Slug)
                ]
                []
            , H.p [ A.class "text-xs text-gray-500 mt-1" ] [ H.text "Leave empty to generate a random slug" ]
            ]
        , H.div
            [ A.class "flex justify-end" ]
            [ viewSubmitButton model ]
        ]


viewSubmitButton : Model -> Html Msg
viewSubmitButton model =
    H.button
        [ A.type_ "submit"
        , A.disabled (isFormDisabled model)
        , A.class
            (if isFormDisabled model then
                "px-6 py-2 bg-gray-300 text-gray-500 rounded-md cursor-not-allowed transition-colors"

             else
                "px-6 py-2 bg-black text-white rounded-md hover:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2 transition-colors"
            )
        ]
        [ H.text "Create note" ]


isFormDisabled : Model -> Bool
isFormDisabled model =
    String.isEmpty model.content


secretUrl : String -> String -> String
secretUrl appUrl slug =
    appUrl ++ "/secret/" ++ slug


viewNoteCreated : Bool -> String -> String -> Html Msg
viewNoteCreated userClickedCopyLink appUrl slug =
    H.div [ A.class "bg-green-50 border border-green-200 rounded-md p-6" ]
        [ H.div [ A.class "bg-white border border-green-300 rounded-md p-4 mb-4" ]
            [ H.p [ A.class "text-sm text-gray-600 mb-2" ]
                [ H.text "Your paste is available at:" ]
            , H.p [ A.class "font-mono text-sm text-gray-800 break-all" ]
                [ H.text (secretUrl appUrl slug) ]
            ]
        , H.div [ A.class "flex gap-3" ]
            [ viewCopyLinkButton userClickedCopyLink
            , viewCreateNewNoteButton
            ]
        ]


viewCopyLinkButton : Bool -> Html Msg
viewCopyLinkButton isClicked =
    let
        base : String
        base =
            "px-4 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2 transition-colors"
    in
    H.button
        [ A.class
            (if isClicked then
                base ++ " bg-green-100 border-green-300 text-green-700"

             else
                base ++ " border-gray-300 text-gray-700 hover:bg-gray-50"
            )
        , E.onClick UserClickedCopyLink
        ]
        [ H.text
            (if isClicked then
                "Copied!"

             else
                "Copy URL"
            )
        ]


viewCreateNewNoteButton : Html Msg
viewCreateNewNoteButton =
    H.button
        [ A.class "px-6 py-2 bg-black text-white rounded-md hover:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2 transition-colors"
        , E.onClick UserClickedCreateNewNote
        ]
        [ H.text "Create New Paste" ]
