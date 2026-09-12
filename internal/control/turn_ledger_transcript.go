package control

import (
	"reasonix/internal/agent"
	slog "reasonix/internal/compat/xslog"
	"reasonix/internal/provider"
	"reasonix/internal/turnevent"
)

func (c *Controller) updateTurnLedgerTranscript(ledger *turnevent.Ledger) *provider.ReadCompletion {
	if c.executor != nil && c.executor.Session() != nil {
		session := c.executor.Session()
		messages, _, rewrite := session.DisplayBaseline()
		ledger.SetTranscriptRewriteEpoch(rewrite)
		digest, digestErr := session.ContentDigest()
		if digestErr != nil {
			slog.Warn("controller: compute terminal transcript digest", "err", digestErr)
		} else {
			ledger.SetTranscriptSnapshot(int64(session.TranscriptVersion()), digest)
		}
		if ref, ok := session.Head(); ok {
			ledger.SetTranscriptHead(ref.HeadID, session.LeafID())
		} else {
			ledger.SetTranscriptHead("", "")
		}
		_rev1 := messages
		for _ri1 := len(_rev1) - 1; _ri1 >= 0; _ri1-- {
			message := _rev1[_ri1]
			if message.ReadCompletion != nil {
				return message.ReadCompletion
			}
			if agent.IsUserAuthoredTurnMessage(message) {
				break
			}
		}
	}
	return nil
}
