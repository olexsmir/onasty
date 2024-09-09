package mailer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMailer_Send(t *testing.T) {
	m := NewTestMailer()
	assert.Empty(t, m.emails)

	email := "test@mail.com"
	err := m.Send(context.TODO(), email, "", "content")
	require.NoError(t, err)

	assert.Equal(t, "content", m.emails[email])
}

func TestMailer_GetLastSentEmailToEmail(t *testing.T) {
	m := NewTestMailer()
	assert.Empty(t, m.emails)

	email := "test@mail.com"
	content := "content"
	err := m.Send(context.TODO(), email, "", content)
	require.NoError(t, err)

	c := m.GetLastSentEmailToEmail(email)
	assert.Equal(t, content, c)
}
