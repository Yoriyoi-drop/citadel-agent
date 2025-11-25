import React from 'react';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';

interface ConfigFieldProps {
    label: string;
    value: string | number;
    onChange: (value: string) => void;
    type?: 'text' | 'number' | 'textarea';
    placeholder?: string;
    required?: boolean;
}

export function ConfigField({
    label,
    value,
    onChange,
    type = 'text',
    placeholder,
    required = false,
}: ConfigFieldProps) {
    return (
        <div className="space-y-2">
            <Label>
                {label}
                {required && <span className="text-red-500 ml-1">*</span>}
            </Label>
            {type === 'textarea' ? (
                <Textarea
                    value={value}
                    onChange={(e) => onChange(e.target.value)}
                    placeholder={placeholder}
                    required={required}
                    aria-label={label}
                />
            ) : (
                <Input
                    type={type}
                    value={value}
                    onChange={(e) => onChange(e.target.value)}
                    placeholder={placeholder}
                    required={required}
                    aria-label={label}
                />
            )}
        </div>
    );
}
