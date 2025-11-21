// utils/node-builder.ts
import { NodeDefinition, NodeInterface } from '../interfaces/node';
import { NodeValidator } from '../validators/node-validator';

export class NodeBuilder {
  private definition: NodeDefinition;
  private nodeClass: typeof NodeInterface;

  constructor(nodeClass: typeof NodeInterface, definition: NodeDefinition) {
    const validation = NodeValidator.validateDefinition(definition);
    if (!validation.isValid) {
      throw new Error(`Invalid node definition: ${validation.errors.join(', ')}`);
    }

    this.nodeClass = nodeClass;
    this.definition = definition;
  }

  static create(nodeClass: typeof NodeInterface, definition: NodeDefinition): NodeBuilder {
    return new NodeBuilder(nodeClass, definition);
  }

  validate(): boolean {
    const validation = NodeValidator.validateDefinition(this.definition);
    return validation.isValid;
  }

  getDefinition(): NodeDefinition {
    return this.definition;
  }

  getClass(): typeof NodeInterface {
    return this.nodeClass;
  }

  build(): { node: typeof NodeInterface; definition: NodeDefinition } {
    return {
      node: this.nodeClass,
      definition: this.definition,
    };
  }
}