#!/usr/bin/env node

const { Command } = require('commander');
const { execa } = require('execa');
const chalk = require('chalk');
const fs = require('fs-extra');
const path = require('path');
const ora = require('ora');
const inquirer = require('inquirer');

const program = new Command();

// Check if Docker is available
async function checkDocker() {
  try {
    await execa('docker', ['--version']);
    return true;
  } catch (error) {
    return false;
  }
}

// Check if Docker Compose is available
async function checkDockerCompose() {
  try {
    await execa('docker-compose', ['--version']);
    return true;
  } catch (error) {
    try {
      // Try docker compose (newer version)
      await execa('docker', ['compose', 'version']);
      return true;
    } catch (error2) {
      return false;
    }
  }
}

// Print Citadel Agent banner
function printBanner() {
  console.log(chalk.bold.blue(`
   █████╗ ██████╗ ██████╗ 
  ██╔══██╗██╔══██╗██╔══██╗
  ███████║██████╔╝██████╔╝
  ██╔══██║██╔═══╝ ██╔═══╝ 
  ██║  ██║██║     ██║     
  ╚═╝  ╚═╝╚═╝     ╚═╝     
  
  Citadel Agent v0.1.0
  Enterprise Workflow Automation Platform
  `));
}

// Install command
program
  .command('install')
  .description('Install Citadel Agent')
  .action(async () => {
    printBanner();
    
    const dockerAvailable = await checkDocker();
    const composeAvailable = await checkDockerCompose();
    
    if (!dockerAvailable || !composeAvailable) {
      console.error(chalk.red('Error: Docker and Docker Compose are required to run Citadel Agent'));
      console.log(chalk.yellow('Please install Docker and Docker Compose first'));
      return;
    }
    
    const spinner = ora('Installing Citadel Agent...');
    spinner.start();
    
    try {
      // Create citadel directory in user's home
      const homeDir = require('os').homedir();
      const citadelDir = path.join(homeDir, '.citadel-agent');
      
      if (!fs.existsSync(citadelDir)) {
        fs.mkdirSync(citadelDir, { recursive: true });
      }
      
      // Copy docker-compose.yml and .env files to the citadel directory
      const projectDir = path.join(__dirname, '..');
      const dockerDir = path.join(projectDir, 'docker');
      
      if (fs.existsSync(dockerDir)) {
        const composeSrc = path.join(dockerDir, 'docker-compose.yml');
        const composeDest = path.join(citadelDir, 'docker-compose.yml');
        
        if (fs.existsSync(composeSrc)) {
          fs.copyFileSync(composeSrc, composeDest);
        }
        
        // Create .env file if not exists
        const envFile = path.join(citadelDir, '.env');
        if (!fs.existsSync(envFile)) {
          fs.writeFileSync(envFile, `# Citadel Agent Configuration
SERVER_PORT=5001
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=citadel_agent
REDIS_ADDR=localhost:6379
JWT_SECRET=your-super-secret-jwt-key-here-change-in-production
JWT_EXPIRY=86400
`);
        }
        
        spinner.succeed('Installation completed successfully!');
        console.log(chalk.green(`\nCitadel Agent installed in: ${citadelDir}`));
        console.log(chalk.yellow(`\nTo start Citadel Agent, run: ${chalk.bold('citadel start')}`));
      } else {
        spinner.fail('Docker directory not found in project');
      }
    } catch (error) {
      spinner.fail('Installation failed');
      console.error(chalk.red(error.message));
    }
  });

// Start command
program
  .command('start')
  .description('Start Citadel Agent services')
  .action(async () => {
    printBanner();
    
    const dockerAvailable = await checkDocker();
    const composeAvailable = await checkDockerCompose();
    
    if (!dockerAvailable || !composeAvailable) {
      console.error(chalk.red('Error: Docker and Docker Compose are required to run Citadel Agent'));
      return;
    }
    
    const homeDir = require('os').homedir();
    const citadelDir = path.join(homeDir, '.citadel-agent');
    
    if (!fs.existsSync(citadelDir)) {
      console.error(chalk.red('Error: Citadel Agent not installed. Run "citadel install" first.'));
      return;
    }
    
    const composeFile = path.join(citadelDir, 'docker-compose.yml');
    if (!fs.existsSync(composeFile)) {
      console.error(chalk.red('Error: docker-compose.yml not found. Run "citadel install" to reinstall.'));
      return;
    }
    
    console.log(chalk.green('Starting Citadel Agent services...'));
    console.log(chalk.yellow('This may take a few moments...'));
    
    try {
      await execa('docker-compose', ['-f', composeFile, 'up', '-d'], {
        cwd: citadelDir,
        stdio: 'inherit'
      });
      
      console.log(chalk.green('\n✓ Citadel Agent is now running!'));
      console.log(chalk.blue('✓ API Server: http://localhost:5001'));
      console.log(chalk.blue('✓ UI: http://localhost:3000 (if UI exists)'));
      console.log(chalk.yellow('\nTo view logs: docker-compose -f ~/.citadel-agent/docker-compose.yml logs -f'));
      console.log(chalk.yellow('To stop: citadel stop'));
    } catch (error) {
      console.error(chalk.red('\nError starting services:'), error.message);
    }
  });

// Stop command
program
  .command('stop')
  .description('Stop Citadel Agent services')
  .action(async () => {
    printBanner();
    
    const dockerAvailable = await checkDocker();
    const composeAvailable = await checkDockerCompose();
    
    if (!dockerAvailable || !composeAvailable) {
      console.error(chalk.red('Error: Docker and Docker Compose are required to run Citadel Agent'));
      return;
    }
    
    const homeDir = require('os').homedir();
    const citadelDir = path.join(homeDir, '.citadel-agent');
    
    if (!fs.existsSync(citadelDir)) {
      console.error(chalk.red('Error: Citadel Agent not installed. Run "citadel install" first.'));
      return;
    }
    
    const composeFile = path.join(citadelDir, 'docker-compose.yml');
    if (!fs.existsSync(composeFile)) {
      console.error(chalk.red('Error: docker-compose.yml not found.'));
      return;
    }
    
    try {
      await execa('docker-compose', ['-f', composeFile, 'down'], {
        cwd: citadelDir,
        stdio: 'inherit'
      });
      
      console.log(chalk.green('✓ Citadel Agent services stopped successfully!'));
    } catch (error) {
      console.error(chalk.red('Error stopping services:'), error.message);
    }
  });

// Reset command
program
  .command('reset')
  .description('Reset Citadel Agent data')
  .action(async () => {
    printBanner();
    
    const answer = await inquirer.prompt([
      {
        type: 'confirm',
        name: 'confirm',
        message: chalk.red('⚠️  This will delete all your workflows and data. Are you sure?'),
        default: false
      }
    ]);
    
    if (!answer.confirm) {
      console.log(chalk.yellow('Operation cancelled'));
      return;
    }
    
    printBanner();
    
    const dockerAvailable = await checkDocker();
    const composeAvailable = await checkDockerCompose();
    
    if (!dockerAvailable || !composeAvailable) {
      console.error(chalk.red('Error: Docker and Docker Compose are required to run Citadel Agent'));
      return;
    }
    
    const homeDir = require('os').homedir();
    const citadelDir = path.join(homeDir, '.citadel-agent');
    
    if (!fs.existsSync(citadelDir)) {
      console.error(chalk.red('Error: Citadel Agent not installed. Run "citadel install" first.'));
      return;
    }
    
    const composeFile = path.join(citadelDir, 'docker-compose.yml');
    if (!fs.existsSync(composeFile)) {
      console.error(chalk.red('Error: docker-compose.yml not found.'));
      return;
    }
    
    try {
      await execa('docker-compose', ['-f', composeFile, 'down', '-v'], {
        cwd: citadelDir,
        stdio: 'inherit'
      });
      
      console.log(chalk.green('✓ Citadel Agent data reset successfully!'));
      console.log(chalk.yellow('To start again: citadel start'));
    } catch (error) {
      console.error(chalk.red('Error resetting data:'), error.message);
    }
  });

// Status command
program
  .command('status')
  .description('Check status of Citadel Agent services')
  .action(async () => {
    printBanner();
    
    const dockerAvailable = await checkDocker();
    const composeAvailable = await checkDockerCompose();
    
    if (!dockerAvailable || !composeAvailable) {
      console.error(chalk.red('Error: Docker and Docker Compose are required to run Citadel Agent'));
      return;
    }
    
    const homeDir = require('os').homedir();
    const citadelDir = path.join(homeDir, '.citadel-agent');
    
    if (!fs.existsSync(citadelDir)) {
      console.log(chalk.red('Citadel Agent is not installed. Run "citadel install" first.'));
      return;
    }
    
    const composeFile = path.join(citadelDir, 'docker-compose.yml');
    if (!fs.existsSync(composeFile)) {
      console.log(chalk.red('docker-compose.yml not found.'));
      return;
    }
    
    try {
      const { stdout } = await execa('docker-compose', ['-f', composeFile, 'ps'], {
        cwd: citadelDir
      });
      
      console.log(stdout || chalk.yellow('No services running'));
    } catch (error) {
      console.error(chalk.red('Error checking status:'), error.message);
    }
  });

// Version command
program
  .command('version')
  .alias('v')
  .description('Show version information')
  .action(() => {
    const packageJson = require('../package.json');
    console.log(`${packageJson.name}/${packageJson.version} ${process.platform}-${process.arch} node-${process.version}`);
  });

// Help command
program
  .name('citadel')
  .description('Citadel Agent - Enterprise Workflow Automation Platform')
  .version(require('./../package.json').version);

program.parse();