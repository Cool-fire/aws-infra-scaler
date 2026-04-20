/**
 * save-memory.ts — Save memory tool (OpenCode 1.4.x plugin)
 *
 * The tool posts memory payloads to a runner-hosted HTTP endpoint specified
 * via `CODEPILOT_MEMORY_URL`. This avoids the process-boundary problem where
 * the opencode subprocess can't share in-memory state with the runner.
 */
import { tool } from '@opencode-ai/plugin';
const CODEPILOT_REPO_URL = process.env.CODEPILOT_REPO_URL ?? '';
const CODEPILOT_MEMORY_URL = process.env.CODEPILOT_MEMORY_URL ?? '';
async function saveMemory(args) {
    if (!CODEPILOT_MEMORY_URL) {
        console.warn('[save-memory] CODEPILOT_MEMORY_URL not set — memory not saved');
        return 'Memory could not be saved (no memory endpoint configured)';
    }
    try {
        const res = await fetch(CODEPILOT_MEMORY_URL, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ repoUrl: CODEPILOT_REPO_URL, ...args }),
        });
        if (!res.ok)
            throw new Error(`memory endpoint ${res.status}`);
    }
    catch (err) {
        console.warn(`[save-memory] failed to send memory: ${String(err)}`);
        return 'Memory could not be saved (transport error)';
    }
    return `Memory saved (${args.memoryType})`;
}
const SaveMemoryPlugin = async () => {
    return {
        tool: {
            save_memory: tool({
                description: 'Save a learning or pattern for future sessions on this repository',
                args: {
                    memoryType: tool.schema
                        .enum(['pattern', 'preference', 'context', 'lesson'])
                        .describe('Type of memory: pattern, preference, context, or lesson'),
                    content: tool.schema.string().describe('The memory content — concise and actionable'),
                },
                async execute(args) {
                    return saveMemory(args);
                },
            }),
        },
    };
};
export default { id: 'codepilot.save-memory', server: SaveMemoryPlugin };
//# sourceMappingURL=save-memory.js.map