#!/usr/bin/env node

const { exec } = require('child_process');
const util = require('util');
const execPromise = util.promisify(exec);

const PORT = process.env.PORT || 3000;

async function killPort(port) {
    console.log(`🔍 Checking for processes on port ${port}...`);

    try {
        // Try to find process using the port
        const { stdout } = await execPromise(`lsof -ti:${port}`);
        const pids = stdout.trim().split('\n').filter(Boolean);

        if (pids.length > 0) {
            console.log(`⚠️  Found ${pids.length} process(es) using port ${port}`);

            // Kill each process
            for (const pid of pids) {
                try {
                    await execPromise(`kill -9 ${pid}`);
                    console.log(`✅ Killed process ${pid}`);
                } catch (err) {
                    console.log(`⚠️  Could not kill process ${pid}: ${err.message}`);
                }
            }

            // Wait a bit for processes to fully terminate
            await new Promise(resolve => setTimeout(resolve, 1000));
            console.log(`✅ Port ${port} is now free`);
        } else {
            console.log(`✅ Port ${port} is already free`);
        }
    } catch (error) {
        // If lsof returns no results, the port is free
        if (error.code === 1) {
            console.log(`✅ Port ${port} is already free`);
        } else {
            console.error(`❌ Error checking port: ${error.message}`);
        }
    }

    // Also kill any lingering Next.js dev processes
    try {
        await execPromise(`pkill -9 -f 'next dev'`);
        console.log(`✅ Killed any lingering Next.js dev processes`);
    } catch (err) {
        // It's okay if there are no processes to kill
    }
}

killPort(PORT).catch(err => {
    console.error('❌ Fatal error:', err);
    process.exit(1);
});
