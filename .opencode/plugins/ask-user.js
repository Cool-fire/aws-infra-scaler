/**
 * ask-user.ts — Ask-user tool (OpenCode 1.4.x plugin)
 *
 * Lets the agent pause mid-turn to ask the user a clarifying question. The
 * tool POSTs to a runner-hosted HTTP endpoint (CODEPILOT_ASK_USER_URL); the
 * runner forwards the question to the control plane over its WebSocket, waits
 * for the client's answer, and resolves the HTTP response with it. The tool
 * returns that answer string as its result so the model can use it.
 */
import { tool } from '@opencode-ai/plugin';
async function askUser(args) {
    const endpoint = process.env.CODEPILOT_ASK_USER_URL ?? '';
    if (!endpoint) {
        return 'Cannot ask the user: CODEPILOT_ASK_USER_URL not configured.';
    }
    const questionId = globalThis.crypto?.randomUUID?.() ??
        `q_${Date.now()}_${Math.random().toString(36).slice(2)}`;
    try {
        const res = await fetch(endpoint, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                questionId,
                question: args.question,
                options: args.options,
            }),
        });
        if (!res.ok)
            throw new Error(`ask-user endpoint ${res.status}`);
        const { answer } = (await res.json());
        return answer;
    }
    catch (err) {
        return `Could not reach the user (transport error: ${String(err)})`;
    }
}
const AskUserPlugin = async () => {
    return {
        tool: {
            ask_user: tool({
                description: 'Ask the end user a clarifying question and wait for their answer. ' +
                    'Use this when you need a decision, preference, or missing detail that ' +
                    'you cannot safely infer. Returns the user\'s answer as a string.',
                args: {
                    question: tool.schema.string().describe('The question to show the user'),
                    options: tool.schema
                        .array(tool.schema.string())
                        .optional()
                        .describe('Optional list of suggested answers to present as choices'),
                },
                async execute(args) {
                    return askUser(args);
                },
            }),
        },
    };
};
export default { id: 'codepilot.ask-user', server: AskUserPlugin };
//# sourceMappingURL=ask-user.js.map