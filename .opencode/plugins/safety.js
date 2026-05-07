/**
 * safety.ts — Safety guardrail plugin (OpenCode 1.4.x)
 *
 * Uses the `tool.execute.before` hook to inspect tool args and throw on
 * dangerous shell commands or writes to protected paths. Throwing inside
 * the hook surfaces as a tool failure to the agent.
 */
const AUTO_DENY_COMMANDS = [
    /rm\s+-rf\s+[~/]/,
    />\s*\/etc\//,
    />\s*\/usr\//,
    /git\s+push\s+--force/,
    /curl[^|]*\|\s*(?:ba)?sh/,
    /wget[^|]*\|\s*(?:ba)?sh/,
];
const AUTO_DENY_PATHS = [
    /\.env$/,
    /\.env\./,
    /\.github\/workflows\//,
    /^\/etc\//,
    /^\/usr\//,
];
const AUTO_APPROVE_COMMANDS = [
    /^(?:grep|find|ls|cat|head|tail|wc|diff|echo|printf)\s/,
    /^(?:npm|pnpm|yarn)\s+(?:test|run\s+test|vitest|jest)/,
    /^(?:npm|pnpm|yarn)\s+install/,
    /^(?:pytest|python\s+-m\s+pytest)/,
];
function loadConfig() {
    try {
        const raw = process.env.CODEPILOT_SAFETY_CONFIG;
        return raw ? JSON.parse(raw) : {};
    }
    catch {
        return {};
    }
}
function checkToolCall(ctx) {
    const config = loadConfig();
    switch (ctx.tool) {
        case 'bash':
        case 'shell':
        case 'run_command':
            return checkShellCommand(String(ctx.args.command ?? ctx.args.cmd ?? ''), config);
        case 'write_file':
        case 'create_file':
        case 'edit_file':
        case 'str_replace_editor':
        case 'write':
        case 'edit':
            return checkFileWrite(String(ctx.args.path ?? ctx.args.file_path ?? ''), config);
        default:
            return { allowed: true };
    }
}
function checkShellCommand(command, config) {
    for (const p of AUTO_APPROVE_COMMANDS)
        if (p.test(command))
            return { allowed: true };
    for (const p of AUTO_DENY_COMMANDS) {
        if (p.test(command))
            return { allowed: false, reason: `Blocked shell command: ${command.slice(0, 80)}` };
    }
    for (const denyPattern of config.denyCommands ?? []) {
        if (command.includes(denyPattern)) {
            return { allowed: false, reason: `Command matches deny list pattern: ${denyPattern}` };
        }
    }
    return { allowed: true };
}
function checkFileWrite(filePath, config) {
    if (config.allowPaths?.some((p) => filePath.startsWith(p)))
        return { allowed: true };
    for (const p of AUTO_DENY_PATHS) {
        if (p.test(filePath))
            return { allowed: false, reason: `Write to protected path: ${filePath}` };
    }
    if (config.denyPaths?.some((p) => filePath.startsWith(p))) {
        return { allowed: false, reason: `Path matches deny list: ${filePath}` };
    }
    return { allowed: true };
}
const SafetyPlugin = async () => {
    return {
        'tool.execute.before': async (input, output) => {
            const verdict = checkToolCall({ tool: input.tool, args: output.args ?? {} });
            if (!verdict.allowed) {
                console.warn(`[safety] blocked tool '${input.tool}': ${verdict.reason}`);
                throw new Error(`Safety guardrail: ${verdict.reason}`);
            }
        },
    };
};
export default { id: 'codepilot.safety', server: SafetyPlugin };
//# sourceMappingURL=safety.js.map