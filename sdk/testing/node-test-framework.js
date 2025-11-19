// citadel-agent/sdk/testing/node-test-framework.js
/**
 * Node Testing Framework for Citadel Agent
 * Provides tools for testing custom nodes
 */

const { NodeFactory } = require('../core/node-factory');
const { NodeValidator } = require('../core/node-validator');
const { NodeManifest } = require('../core/node-manifest');

class NodeTestFramework {
  constructor(options = {}) {
    this.nodeFactory = new NodeFactory();
    this.validator = new NodeValidator();
    this.manifest = new NodeManifest(options.nodesPath || './nodes');
    this.testResults = [];
    this.currentTestId = 0;
    this.testSuites = {};
  }

  /**
   * Creates a test suite for a node
   */
  createTestSuite(nodeId, nodeConfig = {}) {
    const testSuite = {
      id: `test-suite-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
      nodeId,
      config: nodeConfig,
      tests: [],
      startTime: Date.now(),
      results: null
    };

    this.testSuites = this.testSuites || {};
    this.testSuites[testSuite.id] = testSuite;

    return testSuite.id;
  }

  /**
   * Adds a test case to a test suite
   */
  addTestCase(suiteId, testCase) {
    if (!this.testSuites[suiteId]) {
      throw new Error(\`Test suite \${suiteId} does not exist\`);
    }

    const testId = \`test-\${this.currentTestId++}\`;
    const test = {
      id: testId,
      suiteId,
      name: testCase.name || \`Test \${testId}\`,
      description: testCase.description || '',
      input: testCase.input || {},
      expectedOutput: testCase.expectedOutput || {},
      expectedError: testCase.expectedError,
      timeout: testCase.timeout || 10000, // 10 seconds default
      run: testCase.run || (() => {}),
      createdAt: new Date().toISOString()
    };

    this.testSuites[suiteId].tests.push(test);
    return testId;
  }

  /**
   * Runs a single test case
   */
  async runTestCase(testId, suiteId) {
    const testCase = this.testSuites[suiteId]?.tests?.find(t => t.id === testId);
    if (!testCase) {
      throw new Error(\`Test case \${testId} not found in suite \${suiteId}\`);
    }

    const startTime = Date.now();
    let result;
    
    try {
      // Create node instance
      const nodeInstance = this.nodeFactory.createNode(testCase.nodeId, testCase.config);
      
      // Execute test with timeout
      const timeoutPromise = new Promise((_, reject) => {
        setTimeout(() => reject(new Error('Test timeout')), testCase.timeout);
      });
      
      const executionPromise = nodeInstance.execute(testCase.input);
      
      let output;
      try {
        output = await Promise.race([executionPromise, timeoutPromise]);
      } catch (error) {
        output = { status: 'error', error: error.message };
      }

      // Validate output
      result = {
        id: testId,
        status: 'completed',
        input: testCase.input,
        output: output,
        expectedOutput: testCase.expectedOutput,
        expectedError: testCase.expectedError,
        passed: this.validateOutput(output, testCase),
        executionTime: Date.now() - startTime,
        timestamp: new Date().toISOString()
      };
    } catch (error) {
      result = {
        id: testId,
        status: 'error',
        error: error.message,
        executionTime: Date.now() - startTime,
        timestamp: new Date().toISOString()
      };
    }

    // Store result
    testCase.result = result;
    this.testResults.push(result);

    return result;
  }

  /**
   * Validates test output against expectations
   */
  validateOutput(output, testCase) {
    // If expecting an error
    if (testCase.expectedError) {
      if (output.status !== 'error') {
        return false;
      }
      
      if (typeof testCase.expectedError === 'string') {
        return output.error && output.error.includes(testCase.expectedError);
      } else if (testCase.expectedError instanceof RegExp) {
        return output.error && testCase.expectedError.test(output.error);
      } else if (typeof testCase.expectedError === 'function') {
        return testCase.expectedError(output.error);
      } else {
        return true; // Just check that there was an error
      }
    }

    // If expecting success
    if (testCase.expectedOutput) {
      // Simple deep equality check for output.data
      if (output.status === 'success' && testCase.expectedOutput.data) {
        return this.deepEquals(output.data, testCase.expectedOutput.data);
      }
      
      // Check status if specified
      if (testCase.expectedOutput.status) {
        return output.status === testCase.expectedOutput.status;
      }
      
      // If no specific expectations, just check for success status
      return output.status === 'success';
    }

    // If no expectations, consider it passed if there's no error
    return output.status !== 'error';
  }

  /**
   * Simple deep equals implementation for basic objects
   */
  deepEquals(obj1, obj2) {
    if (obj1 === obj2) return true;
    if (obj1 == null || obj2 == null) return false;
    if (typeof obj1 !== 'object' || typeof obj2 !== 'object') return obj1 === obj2;

    const keys1 = Object.keys(obj1);
    const keys2 = Object.keys(obj2);
    
    if (keys1.length !== keys2.length) return false;
    
    for (const key of keys1) {
      if (!keys2.includes(key)) return false;
      if (!this.deepEquals(obj1[key], obj2[key])) return false;
    }
    
    return true;
  }

  /**
   * Runs all tests in a suite
   */
  async runTestSuite(suiteId) {
    const testSuite = this.testSuites[suiteId];
    if (!testSuite) {
      throw new Error(\`Test suite \${suiteId} does not exist\`);
    }

    testSuite.startTime = Date.now();

    for (const testCase of testSuite.tests) {
      console.log(\`Running test: \${testCase.name}\`);
      await this.runTestCase(testCase.id, suiteId);
    }

    testSuite.endTime = Date.now();
    testSuite.duration = testSuite.endTime - testSuite.startTime;

    // Calculate summary
    const results = testSuite.tests.map(t => t.result);
    const passed = results.filter(r => r.status === 'completed' && r.passed).length;
    const failed = results.filter(r => r.status === 'completed' && !r.passed).length;
    const errors = results.filter(r => r.status === 'error').length;

    testSuite.summary = {
      total: results.length,
      passed,
      failed,
      errors,
      successRate: results.length > 0 ? (passed / results.length) * 100 : 0,
      duration: testSuite.duration
    };

    return testSuite.summary;
  }

  /**
   * Runs multiple test suites
   */
  async runTestSuites(suiteIds) {
    const results = {};
    
    for (const suiteId of suiteIds) {
      results[suiteId] = await this.runTestSuite(suiteId);
    }
    
    return results;
  }

  /**
   * Creates a test harness for a node
   */
  createNodeTestHarness(nodeId, config = {}) {
    const testCases = [];

    return {
      nodeId,
      config,
      testCases,

      addTest: function(name, options) {
        const testCase = {
          name: name,
          description: options.description || '',
          input: options.input || {},
          expectedOutput: options.expectedOutput,
          expectedError: options.expectedError,
          timeout: options.timeout || 10000,
          nodeId: nodeId,
          config: config
        };

        testCases.push(testCase);
        return this;
      },

      testInputOutput: function(input, expectedOutput) {
        return this.addTest('Input/Output Test', {
          input,
          expectedOutput
        });
      },

      testErrorHandling: function(input, expectedError) {
        return this.addTest('Error Handling Test', {
          input,
          expectedError
        });
      },

      testPerformance: function(input, maxExecutionTime) {
        return this.addTest('Performance Test', {
          input,
          timeout: maxExecutionTime * 2, // Set timeout higher for performance tests
          expectedOutput: (result) => {
            // Custom validation function for performance tests
            if (typeof result.metadata?.executionTime !== 'number') {
              return false;
            }
            return result.metadata.executionTime <= maxExecutionTime;
          }
        });
      },

      run: async () => {
        const suiteId = this.createTestSuite(nodeId, config);

        for (const testCase of testCases) {
          this.addTestCase(suiteId, testCase);
        }

        return await this.runTestSuite(suiteId);
      }
    };
  }

  /**
   * Validates a node implementation before testing
   */
  async validateNodeForTesting(nodeClassOrPath) {
    // If it's a class, validate the class directly
    if (typeof nodeClassOrPath === 'function') {
      return this.validator.validateNodeImplementation(nodeClassOrPath);
    }
    
    // If it's a path, load and validate
    if (typeof nodeClassOrPath === 'string') {
      try {
        const nodeModule = require(nodeClassOrPath);
        const NodeClass = nodeModule.default || nodeModule;
        
        if (typeof NodeClass !== 'function') {
          return { valid: false, errors: ['Node module does not export a constructor'] };
        }
        
        return this.validator.validateNodeImplementation(NodeClass);
      } catch (error) {
        return { valid: false, errors: [\`Could not load node: \${error.message}\`] };
      }
    }
    
    return { valid: false, errors: ['Node must be a class or path to a module'] };
  }

  /**
   * Generates test report
   */
  generateReport(suiteId = null) {
    let suitesToReport;
    
    if (suiteId) {
      suitesToReport = { [suiteId]: this.testSuites[suiteId] };
    } else {
      suitesToReport = this.testSuites;
    }

    const report = {
      generatedAt: new Date().toISOString(),
      totals: {
        suites: 0,
        tests: 0,
        passed: 0,
        failed: 0,
        errors: 0
      },
      suites: {}
    };

    for (const [suiteId, suite] of Object.entries(suitesToReport)) {
      if (!suite.summary) {
        continue; // Skip suites that haven't been run
      }

      report.suites[suiteId] = {
        nodeId: suite.nodeId,
        config: suite.config,
        summary: suite.summary,
        tests: suite.tests.map(test => ({
          id: test.id,
          name: test.name,
          description: test.description,
          result: test.result ? {
            status: test.result.status,
            passed: test.result.passed,
            executionTime: test.result.executionTime
          } : null
        }))
      };

      report.totals.suites++;
      report.totals.tests += suite.summary.total;
      report.totals.passed += suite.summary.passed;
      report.totals.failed += suite.summary.failed;
      report.totals.errors += suite.summary.errors;
    }

    report.totals.successRate = report.totals.tests > 0 ? 
      (report.totals.passed / report.totals.tests) * 100 : 0;

    return report;
  }

  /**
   * Saves test report to file
   */
  async saveReport(report, filename = null) {
    const fs = require('fs').promises;
    const path = require('path');
    
    const reportFilename = filename || \`test-report-\${new Date().toISOString().replace(/[:.]/g, '-')}.json\`;
    const reportPath = path.join(process.cwd(), reportFilename);
    
    await fs.writeFile(reportPath, JSON.stringify(report, null, 2));
    console.log(\`Test report saved to: \${reportPath}\`);
    
    return reportPath;
  }

  /**
   * Runs a quick smoke test on a node
   */
  async smokeTest(nodeId, config = {}) {
    const harness = this.createNodeTestHarness(nodeId, config);
    
    // Add basic smoke test
    harness.addTest('Smoke Test', {
      input: { test: 'data' },
      expectedOutput: { status: 'success' }
    });
    
    const result = await harness.run();
    return result;
  }
}

// Convenience function for creating and running a quick test
async function quickNodeTest(nodeId, config = {}, testCases = []) {
  const framework = new NodeTestFramework();
  const harness = framework.createNodeTestHarness(nodeId, config);
  
  for (const testCase of testCases) {
    if (testCase.input && testCase.expectedOutput) {
      harness.testInputOutput(testCase.input, testCase.expectedOutput);
    } else if (testCase.input && testCase.expectedError) {
      harness.addTest('Error Test', {
        input: testCase.input,
        expectedError: testCase.expectedError
      });
    }
  }
  
  return await harness.run();
}

module.exports = {
  NodeTestFramework,
  quickNodeTest
};