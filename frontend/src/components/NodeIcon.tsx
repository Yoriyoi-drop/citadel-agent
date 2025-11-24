import React from 'react';
import type { LucideIcon } from 'lucide-react';
import { HelpCircle } from 'lucide-react';
import { getNodeIcon, getCategoryIcon, getCategoryColor } from '@/config/nodeIcons';

interface NodeIconProps {
    /** Node type identifier */
    type: string;
    /** Icon size in pixels */
    size?: number;
    /** Icon color (hex or CSS color) */
    color?: string;
    /** Additional CSS classes */
    className?: string;
    /** Icon stroke width */
    strokeWidth?: number;
}

/**
 * NodeIcon Component
 * Renders the appropriate icon for a given node type
 */
export const NodeIcon: React.FC<NodeIconProps> = ({
    type,
    size = 20,
    color,
    className = '',
    strokeWidth = 2,
}) => {
    const iconComponent = getNodeIcon(type);

    // Handle both Lucide icons and brand icon strings
    if (typeof iconComponent === 'string') {
        // Brand icon - would need Simple Icons integration
        return (
            <div
                className={`inline-flex items-center justify-center ${className}`}
                style={{ width: size, height: size }}
            >
                <span className="text-xs font-bold">{iconComponent.slice(0, 2).toUpperCase()}</span>
            </div>
        );
    }

    const Icon = iconComponent as LucideIcon;

    return (
        <Icon
            size={size}
            color={color}
            strokeWidth={strokeWidth}
            className={className}
        />
    );
};

interface CategoryIconProps {
    /** Category identifier */
    category: string;
    /** Icon size in pixels */
    size?: number;
    /** Use category color */
    useColor?: boolean;
    /** Additional CSS classes */
    className?: string;
    /** Icon stroke width */
    strokeWidth?: number;
}

/**
 * CategoryIcon Component
 * Renders the icon for a node category
 */
export const CategoryIcon: React.FC<CategoryIconProps> = ({
    category,
    size = 20,
    useColor = true,
    className = '',
    strokeWidth = 2,
}) => {
    const Icon = getCategoryIcon(category);
    const color = useColor ? getCategoryColor(category) : undefined;

    return (
        <Icon
            size={size}
            color={color}
            strokeWidth={strokeWidth}
            className={className}
        />
    );
};

interface NodeIconBadgeProps {
    /** Node type identifier */
    type: string;
    /** Category for color */
    category?: string;
    /** Icon size in pixels */
    size?: number;
    /** Show background */
    showBackground?: boolean;
    /** Additional CSS classes */
    className?: string;
}

/**
 * NodeIconBadge Component
 * Renders a node icon with optional colored background
 */
export const NodeIconBadge: React.FC<NodeIconBadgeProps> = ({
    type,
    category,
    size = 24,
    showBackground = true,
    className = '',
}) => {
    const color = category ? getCategoryColor(category) : '#64748b';

    return (
        <div
            className={`inline-flex items-center justify-center rounded-lg ${className}`}
            style={{
                width: size * 1.5,
                height: size * 1.5,
                backgroundColor: showBackground ? `${color}20` : 'transparent',
            }}
        >
            <NodeIcon
                type={type}
                size={size}
                color={color}
                strokeWidth={2}
            />
        </div>
    );
};

export default NodeIcon;
