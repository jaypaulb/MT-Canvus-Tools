package qa

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/ai-personas/internal/ai"
	"github.com/jaypaulb/MT-Canvus-Tools/go/examples/projects/ai-personas/internal/atom"
)

// MinAnswers is the smallest number of successful answers that allows the
// workflow to keep going (the rest of the layout copes with sparse results).
const MinAnswers = 1

// generateAnswers asks every persona the question in parallel, applying a
// "rephrase succinctly" follow-up if the result exceeds tokenLimit characters.
// Indices of failed personas hold an empty string in the returned slice; the
// errors slice carries diagnostic context for each failure.
func generateAnswers(ctx context.Context, sm *ai.SessionManager, personas []atom.Persona, question, businessContext string, tokenLimit int) ([]string, []error) {
	answers := make([]string, len(personas))
	errs := make([]error, len(personas))
	var wg sync.WaitGroup
	wg.Add(len(personas))
	for i, p := range personas {
		go func(i int, p atom.Persona) {
			defer wg.Done()
			ans, err := sm.AnswerAs(ctx, p, businessContext, question)
			if err != nil {
				errs[i] = fmt.Errorf("persona %s: %w", p.Name, err)
				return
			}
			if len(ans) > tokenLimit {
				succinct := fmt.Sprintf(
					"Please rephrase your answer in a much more succinct, short, and verbal way. Limit your response to %d characters.",
					tokenLimit)
				ans2, err := sm.AnswerAs(ctx, p, businessContext, succinct)
				if err != nil {
					errs[i] = fmt.Errorf("persona %s (succinct): %w", p.Name, err)
					return
				}
				ans = ans2
			}
			answers[i] = ans
		}(i, p)
	}
	wg.Wait()
	return answers, errs
}

// generateMetaAnswers re-asks each persona how the other personas' answers
// changed their thinking. Same return contract as generateAnswers; personas
// whose initial answer is empty/errored are skipped.
func generateMetaAnswers(ctx context.Context, sm *ai.SessionManager, personas []atom.Persona, originalAnswers []string, answerErrs []error, businessContext string, tokenLimit int) ([]string, []error) {
	out := make([]string, len(personas))
	errs := make([]error, len(personas))
	var wg sync.WaitGroup
	for i, p := range personas {
		if originalAnswers[i] == "" || answerErrs[i] != nil {
			continue
		}
		var others []string
		for j, ans := range originalAnswers {
			if i != j && ans != "" && answerErrs[j] == nil {
				others = append(others, fmt.Sprintf("%s said: %s", personas[j].Name, ans))
			}
		}
		if len(others) == 0 {
			out[i] = "No other responses to react to."
			continue
		}
		wg.Add(1)
		go func(i int, p atom.Persona, others []string) {
			defer wg.Done()
			prompt := fmt.Sprintf("Thank you %s for the interesting answer. Does what you heard from the others change what you think in any way? You heard: %s",
				p.Name, joinSemi(others))
			ans, err := sm.AnswerAs(ctx, p, businessContext, prompt)
			if err != nil {
				errs[i] = fmt.Errorf("persona %s meta: %w", p.Name, err)
				return
			}
			if len(ans) > tokenLimit {
				succinct := fmt.Sprintf(
					"Please rephrase your answer in a much more succinct, short, and verbal way. Limit your response to %d characters.",
					tokenLimit)
				ans2, err := sm.AnswerAs(ctx, p, businessContext, succinct)
				if err != nil {
					errs[i] = fmt.Errorf("persona %s meta (succinct): %w", p.Name, err)
					return
				}
				ans = ans2
			}
			out[i] = ans
		}(i, p, others)
	}
	wg.Wait()
	return out, errs
}

// joinSemi joins parts with "; " separators (matching the legacy prompt
// format that meta-answer Q&A consumers may have been trained against).
func joinSemi(parts []string) string {
	switch len(parts) {
	case 0:
		return ""
	case 1:
		return parts[0]
	}
	out := parts[0]
	for _, p := range parts[1:] {
		out += "; " + p
	}
	return out
}

// countSuccess returns the number of non-empty entries whose corresponding
// error is nil.
func countSuccess(answers []string, errs []error) int {
	n := 0
	for i := range answers {
		if answers[i] != "" && errs[i] == nil {
			n++
		}
	}
	return n
}

// logAnswerStats writes a structured summary of generation outcomes.
func logAnswerStats(label string, answers []string, errs []error) {
	n := countSuccess(answers, errs)
	if n < len(answers) {
		slog.Warn("qa.answer: partial success", "phase", label, "ok", n, "total", len(answers))
	} else {
		slog.Info("qa.answer: complete", "phase", label, "ok", n, "total", len(answers))
	}
}
