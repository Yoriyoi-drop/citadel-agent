// scripts/citadel-agent.js
// Simple Node.js orchestrator for Citadel Agent CLI commands.
// Usage: npx citadel-agent <command>
// Commands implemented: start, dev, build, init, doctor

const { execSync, spawn } = require('child_process');
const fs = require('fs');
const path = require('path');

const PROJECT_ROOT = path.resolve(__dirname, '..');
const BIN_PATH = path.join(PROJECT_ROOT, 'bin', 'citadel-agent');
const BACKEND_MAIN = path.join(PROJECT_ROOT, 'backend', 'main.go');
const CONFIG_FILE = path.join(PROJECT_ROOT, 'config.yaml');
const WORKFLOWS_DIR = path.join(PROJECT_ROOT, 'workflows');

function logSuccess(msg) {
    console.log('\x1b[32m✔\x1b[0m', msg);
}
function logError(msg) {
    console.error('\x1b[31m✖\x1b[0m', msg);
}
function logInfo(msg) {
    console.log('\x1b[36mℹ\x1b[0m', msg);
}
function exec(command, options = {}) {
    try {
        execSync(command, { stdio: 'inherit', ...options });
        return true;
    } catch (e) {
        return false;
    }
}

function commandStart() {
    // Ensure binary exists, otherwise build it first
    if (!fs.existsSync(BIN_PATH)) {
        logInfo('Binary not found, building first...');
        if (!commandBuild()) {
            logError('Build failed, cannot start server');
            process.exit(1);
        }
    }
    logInfo('Starting Citadel Agent backend...');
    const child = spawn(BIN_PATH, [], { stdio: 'inherit', cwd: PROJECT_ROOT });
    child.on('close', code => {
        logInfo(`Backend process exited with code ${code}`);
    });
}

function commandDev() {
    // Run backend with go run (auto‑reload not implemented, simple for now)
    logInfo('Running backend in dev mode (go run)...');
    const backend = spawn('go', ['run', BACKEND_MAIN], { stdio: 'inherit', cwd: PROJECT_ROOT });
    // Also start Next.js dev server
    logInfo('Starting Next.js dev server...');
    const frontend = spawn('npm', ['run', 'dev'], { stdio: 'inherit', cwd: PROJECT_ROOT });
    // Forward exit
    backend.on('close', code => logInfo(`Backend exited with ${code}`));
    frontend.on('close', code => logInfo(`Frontend exited with ${code}`));
}

function commandBuild() {
    // Build Go binary to ./bin/citadel-agent
    const binDir = path.dirname(BIN_PATH);
    if (!fs.existsSync(binDir)) fs.mkdirSync(binDir, { recursive: true });
    logInfo('Building Go binary...');
    const ok = exec(`go build -o ${BIN_PATH} ${BACKEND_MAIN}`);
    if (ok) logSuccess('Binary built successfully');
    else logError('Go build failed');
    return ok;
}

function commandInit() {
    // Create config.yaml if missing
    if (!fs.existsSync(CONFIG_FILE)) {
        const defaultConfig = `# Citadel Agent configuration\nport: 8080\nlogLevel: info\n`;
        fs.writeFileSync(CONFIG_FILE, defaultConfig);
        logSuccess('Created config.yaml');
    } else {
        logInfo('config.yaml already exists');
    }
    // Ensure workflows folder exists with a sample file
    if (!fs.existsSync(WORKFLOWS_DIR)) {
        fs.mkdirSync(WORKFLOWS_DIR);
        logSuccess('Created workflows directory');
    }
    const sample = path.join(WORKFLOWS_DIR, 'example.json');
    if (!fs.existsSync(sample)) {
        const example = { name: 'example', description: 'sample workflow', nodes: [], edges: [] };
        fs.writeFileSync(sample, JSON.stringify(example, null, 2));
        logSuccess('Created example workflow');
    } else {
        logInfo('example workflow already exists');
    }
}

function commandDoctor() {
    // Go version
    const goVersion = execSync('go version', { encoding: 'utf8' }).trim();
    logSuccess(`Go version: ${goVersion}`);
    // Port check (8080)
    const port = 8080;
    const isPortFree = execSync(`lsof -i:${port} || true`, { encoding: 'utf8' }).trim() === '';
    if (isPortFree) logSuccess(`Port ${port} is free`);
    else logError(`Port ${port} is in use`);
    // Config file existence
    if (fs.existsSync(CONFIG_FILE)) logSuccess('config.yaml exists');
    else logError('config.yaml missing');
    // Binary existence
    if (fs.existsSync(BIN_PATH)) logSuccess('Binary exists');
    else logInfo('Binary not built yet');
}

function main() {
    const args = process.argv.slice(2);
    const cmd = args[0];
    switch (cmd) {
        case 'start':
            commandStart();
            break;
        case 'dev':
            commandDev();
            break;
        case 'build':
            commandBuild();
            break;
        case 'init':
            commandInit();
            break;
        case 'doctor':
            commandDoctor();
            break;
        default:
            console.log('Usage: npx citadel-agent <start|dev|build|init|doctor>');
    }
}

main();
