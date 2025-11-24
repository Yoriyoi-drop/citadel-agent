/**
 * Node Icon Configuration
 * Maps node types to their corresponding icons from Lucide and Simple Icons
 */

import {
    // HTTP & API
    Globe, Cloud, Wifi, Link, Zap, Activity, Send, Webhook,

    // Database
    Database, Server, HardDrive, Archive, Save, RefreshCw,

    // AI - General
    Brain, Sparkles, Cpu, Bot, Wand2,

    // AI - Vision
    Image, Eye, Camera, Scan, Crop, Maximize2,

    // AI - Speech
    Mic, Volume2, Radio, Headphones, Speaker, AudioLines,

    // AI - NLP
    FileText, MessageSquare, Type, Languages, BookOpen, Hash,

    // Data Transform
    Shuffle, Code, Filter, GitMerge, Split, Combine,

    // Validation & Logic
    CheckCircle, AlertCircle, GitBranch, Scale, Shield, Flag,

    // Flow Control
    Repeat, FastForward, Play, Pause, SkipForward, StopCircle,

    // File Operations
    File, Folder, Upload, Download, Trash, FileArchive,

    // Cloud Storage
    CloudUpload, CloudDownload, CloudCog,

    // Communication
    Mail, Send as SendIcon, Bell, Phone, MessageCircle,

    // CRM & Marketing
    Users, TrendingUp, BarChart, PieChart, Target,

    // Social Media
    Share2, Heart, ThumbsUp, Star, Bookmark,

    // Payment
    CreditCard, DollarSign, ShoppingCart, Receipt, Wallet,

    // Scheduling
    Calendar, Clock, Timer, AlarmClock, Hourglass,

    // Security
    Lock, Key, EyeOff, Fingerprint, ShieldCheck,

    // Monitoring
    TrendingDown, AlertTriangle, LineChart, Activity as ActivityIcon,

    // Utilities
    Wrench, Settings, Smartphone, Tablet,

    // Search & RAG
    Search, Layers, Link2, Database as DatabaseIcon,

    // General
    HelpCircle, Info, AlertOctagon, CheckSquare, XCircle,
    Package, Box, Boxes, Grid, List, Table,

    // Additional
    Pencil, Files,
} from 'lucide-react';

import type { LucideIcon } from 'lucide-react';

// Type for icon mapping
export type NodeIconType = LucideIcon | string;

/**
 * Primary icon mapping for all node types
 */
export const nodeIcons: Record<string, NodeIconType> = {
    // ========== HTTP & API (30 nodes) ==========
    http_request: Globe,
    http_get: Globe,
    http_post: Send,
    http_put: RefreshCw,
    http_delete: Trash,
    http_patch: Pencil,
    graphql_query: Code,
    graphql_mutation: Zap,
    rest_api: Cloud,
    soap_request: Cloud,
    webhook_trigger: Webhook,
    webhook_send: Send,
    api_gateway: Shield,
    rate_limiter: Timer,
    api_versioning: GitBranch,

    // ========== Database (40 nodes) ==========
    postgres_query: Database,
    postgres_insert: Save,
    postgres_update: RefreshCw,
    postgres_delete: Trash,
    mysql_query: Database,
    mongodb_find: Database,
    mongodb_insert: Save,
    mongodb_update: RefreshCw,
    mongodb_delete: Trash,
    redis_get: Database,
    redis_set: Save,
    redis_delete: Trash,
    elasticsearch_search: Search,
    elasticsearch_index: Archive,
    dynamodb_query: Database,
    cassandra_query: Database,
    neo4j_query: GitBranch,
    sqlite_query: Database,

    // ========== AI - LLM (35 nodes) ==========
    llm_chat: Brain,
    llm_completion: Sparkles,
    openai_chat: Brain,
    openai_completion: Sparkles,
    claude_chat: Brain,
    gemini_chat: Brain,
    llama_chat: Brain,
    mistral_chat: Brain,
    gpt4_chat: Brain,
    gpt35_chat: Brain,
    text_generation: FileText,
    prompt_template: MessageSquare,
    few_shot_learning: Layers,
    chain_of_thought: GitBranch,
    rag_query: Search,

    // ========== AI - Vision (25 nodes) ==========
    image_classification: Image,
    object_detection: Eye,
    face_recognition: Camera,
    ocr_text_extraction: Scan,
    image_segmentation: Crop,
    image_generation: Wand2,
    stable_diffusion: Image,
    dall_e: Image,
    midjourney: Image,
    image_upscaling: Maximize2,
    background_removal: Crop,

    // ========== AI - Speech (20 nodes) ==========
    speech_to_text: Mic,
    text_to_speech: Speaker,
    whisper_transcribe: Mic,
    voice_cloning: Volume2,
    audio_classification: AudioLines,
    speech_recognition: Mic,
    voice_synthesis: Speaker,

    // ========== AI - NLP (25 nodes) ==========
    text_classification: FileText,
    sentiment_analysis: Heart,
    named_entity_recognition: Hash,
    text_summarization: FileText,
    translation: Languages,
    question_answering: HelpCircle,
    text_embedding: Layers,
    semantic_search: Search,

    // ========== Data Transform (30 nodes) ==========
    json_parse: Code,
    json_stringify: Code,
    xml_parse: Code,
    csv_parse: Table,
    data_mapper: Shuffle,
    data_filter: Filter,
    data_sort: List,
    data_merge: GitMerge,
    data_split: Split,
    data_aggregate: Combine,

    // ========== Validation & Logic (25 nodes) ==========
    if_condition: GitBranch,
    switch_case: GitBranch,
    validate_schema: CheckCircle,
    validate_email: Mail,
    validate_url: Link,
    validate_json: Code,
    regex_match: Hash,
    compare_values: Scale,

    // ========== Flow Control (20 nodes) ==========
    loop_foreach: Repeat,
    loop_while: Repeat,
    delay: Timer,
    wait: Clock,
    retry: RefreshCw,
    timeout: AlarmClock,
    parallel_execution: Zap,
    sequential_execution: List,

    // ========== File Operations (25 nodes) ==========
    read_file: File,
    write_file: Save,
    delete_file: Trash,
    move_file: Shuffle,
    copy_file: Files,
    list_files: Folder,
    compress_file: Archive,
    decompress_file: FileArchive,

    // ========== Cloud Storage (20 nodes) ==========
    s3_upload: CloudUpload,
    s3_download: CloudDownload,
    s3_list: Cloud,
    gcs_upload: CloudUpload,
    azure_blob_upload: CloudUpload,
    dropbox_upload: CloudUpload,
    google_drive_upload: CloudUpload,

    // ========== Communication (30 nodes) ==========
    send_email: Mail,
    send_sms: MessageCircle,
    send_slack: MessageSquare,
    send_telegram: Send,
    send_discord: MessageCircle,
    send_whatsapp: Phone,
    send_notification: Bell,

    // ========== CRM & Marketing (25 nodes) ==========
    salesforce_create: Users,
    salesforce_update: RefreshCw,
    hubspot_create: Users,
    mailchimp_send: Mail,
    sendgrid_send: Mail,

    // ========== Social Media (20 nodes) ==========
    twitter_post: Share2,
    facebook_post: Share2,
    instagram_post: Image,
    linkedin_post: Share2,

    // ========== Payment (25 nodes) ==========
    stripe_charge: CreditCard,
    stripe_refund: DollarSign,
    paypal_payment: Wallet,
    shopify_order: ShoppingCart,

    // ========== Scheduling (20 nodes) ==========
    cron_trigger: Calendar,
    schedule_task: Clock,
    delay_task: Timer,

    // ========== Security (25 nodes) ==========
    encrypt_data: Lock,
    decrypt_data: Key,
    hash_password: Shield,
    verify_signature: ShieldCheck,
    jwt_sign: Key,
    jwt_verify: ShieldCheck,

    // ========== Monitoring (20 nodes) ==========
    log_message: FileText,
    track_metric: BarChart,
    alert_trigger: AlertTriangle,
    health_check: Activity,

    // ========== Utilities (20 nodes) ==========
    random_number: Hash,
    uuid_generate: Hash,
    date_format: Calendar,
    string_concat: Type,
    math_operation: Hash,

    // ========== AI - RAG (20 nodes) ==========
    vector_store: Database,
    embedding_create: Layers,
    similarity_search: Search,
    document_loader: FileText,
    text_splitter: Split,
};

/**
 * Category icon mapping
 */
export const categoryIcons: Record<string, LucideIcon> = {
    http: Globe,
    database: Database,
    ai_llm: Brain,
    ai_vision: Image,
    ai_speech: Mic,
    ai_nlp: FileText,
    transform: Shuffle,
    validation: CheckCircle,
    flow: GitBranch,
    file: File,
    cloud: Cloud,
    communication: MessageSquare,
    crm: Users,
    social: Share2,
    payment: CreditCard,
    schedule: Calendar,
    security: Lock,
    monitoring: Activity,
    utility: Wrench,
    rag: Search,
};

/**
 * Category color mapping (Tailwind colors)
 */
export const categoryColors: Record<string, string> = {
    http: '#3b82f6',        // blue-500
    database: '#10b981',    // green-500
    ai_llm: '#8b5cf6',      // purple-500
    ai_vision: '#ec4899',   // pink-500
    ai_speech: '#f59e0b',   // amber-500
    ai_nlp: '#06b6d4',      // cyan-500
    transform: '#14b8a6',   // teal-500
    validation: '#22c55e',  // green-500
    flow: '#0ea5e9',        // sky-500
    file: '#6366f1',        // indigo-500
    cloud: '#8b5cf6',       // purple-500
    communication: '#ec4899', // pink-500
    crm: '#f97316',         // orange-500
    social: '#ef4444',      // red-500
    payment: '#84cc16',     // lime-500
    schedule: '#06b6d4',    // cyan-500
    security: '#dc2626',    // red-600
    monitoring: '#eab308',  // yellow-500
    utility: '#64748b',     // slate-500
    rag: '#a855f7',         // purple-500
};

/**
 * Brand/Integration icon mapping (using Simple Icons)
 * These are string identifiers that will be loaded from simple-icons
 */
export const brandIcons: Record<string, string> = {
    // Cloud Providers
    aws: 'amazonaws',
    gcp: 'googlecloud',
    azure: 'microsoftazure',

    // Databases
    postgresql: 'postgresql',
    mysql: 'mysql',
    mongodb: 'mongodb',
    redis: 'redis',
    elasticsearch: 'elasticsearch',

    // AI Services
    openai: 'openai',
    anthropic: 'anthropic',
    google_ai: 'google',

    // Communication
    slack: 'slack',
    discord: 'discord',
    telegram: 'telegram',
    whatsapp: 'whatsapp',

    // CRM & Marketing
    salesforce: 'salesforce',
    hubspot: 'hubspot',
    mailchimp: 'mailchimp',
    sendgrid: 'sendgrid',

    // Social Media
    twitter: 'twitter',
    facebook: 'facebook',
    instagram: 'instagram',
    linkedin: 'linkedin',

    // Payment
    stripe: 'stripe',
    paypal: 'paypal',
    shopify: 'shopify',

    // Storage
    dropbox: 'dropbox',
    google_drive: 'googledrive',
    onedrive: 'microsoftonedrive',
};

/**
 * Get icon for a node type
 */
export function getNodeIcon(nodeType: string): NodeIconType {
    return nodeIcons[nodeType] || HelpCircle;
}

/**
 * Get category icon
 */
export function getCategoryIcon(category: string): LucideIcon {
    return categoryIcons[category] || Box;
}

/**
 * Get category color
 */
export function getCategoryColor(category: string): string {
    return categoryColors[category] || '#64748b';
}

/**
 * Get brand icon identifier
 */
export function getBrandIcon(brand: string): string | null {
    return brandIcons[brand] || null;
}
