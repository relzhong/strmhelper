const translations = {
    en: {
        dashboard: "Dashboard",
        logout: "Logout",
        add_task: "Add New Task",
        edit_task: "Edit Strm Task",
        new_task: "New Strm Task",
        save_task: "Save Task",
        run_now: "Run Now",
        edit: "Edit",
        delete: "Delete",
        confirm_delete: "Are you sure?",
        no_tasks: "No tasks found. Click \"Add New Task\" to get started.",
        running: "Running",
        
        // Sections
        basic_info: "Basic Information",
        connection: "Openlist Connection",
        paths: "Paths",
        advanced_settings: "Advanced Settings & Features",
        smart_protection: "Smart Protection",

        // Fields
        task_id: "Task ID",
        task_id_help: "Unique identifier for this task",
        cron: "Cron Expression",
        cron_help: "Scheduled time for background tasks",
        url: "Server URL",
        url_help: "Openlist server address",
        public_url: "Public URL",
        public_url_help: "Publicly accessible address (optional)",
        username: "Username",
        username_help: "Openlist account username",
        password: "Password",
        password_help: "Openlist account password",
        token: "Token",
        token_help: "Permanent access token (optional)",
        source_dir: "Source Directory",
        source_dir_help: "Folder path on the server",
        target_dir: "Target Directory",
        target_dir_help: "Local path for output",
        next_run: "Next Execution",
        last_error: "Last Error",
        flatten_mode: "Flatten Mode",
        flatten_mode_help: "Disables subtitles, images, and NFOs",
        subtitle: "Download Subtitle",
        subtitle_help: "Download subtitle files",
        image: "Download Image",
        image_help: "Download image files",
        nfo: "Download NFO",
        nfo_help: "Download .nfo files",
        mode: "Mode",
        mode_help: "Strm file content format",
        overwrite: "Overwrite",
        overwrite_help: "Re-generate if file exists",
        sync_server: "Sync Server",
        sync_server_help: "Synchronize with server state",
        sync_ignore: "Sync Ignore (Regex)",
        sync_ignore_help: "Regular expression for ignored files",
        sync_back: "Sync Back Delete",
        sync_back_help: "Sync local deletions back to the remote server (CAUTION)",
        other_ext: "Other Extensions",
        other_ext_help: "Custom extensions (comma separated)",
        check_mod_time: "Directory Time Check",
        check_mod_time_help: "Skip scanning if directory modification time hasn't changed (WARNING: May miss deletions on some cloud drives)",
        max_workers: "Max Workers",
        max_workers_help: "Maximum concurrent requests",
        max_downloaders: "Max Downloaders",
        max_downloaders_help: "Maximum simultaneous downloads",
        wait_time: "Wait Time (s)",
        wait_time_help: "Delay between requests",
        enabled: "Enabled",
        enabled_help: "Enable smart protection",
        threshold: "Threshold",
        threshold_help: "Minimum files to trigger protection",
        grace_scans: "Grace Scans",
        grace_scans_help: "Scans required before deletion",

        // Login
        login_title: "StrmHelper Login",
        login_btn: "Sign In",
        invalid_creds: "Invalid credentials"
    },
    zh: {
        dashboard: "控制面板",
        logout: "退出登录",
        add_task: "添加新任务",
        edit_task: "编辑 Strm 任务",
        new_task: "新建 Strm 任务",
        save_task: "保存任务",
        run_now: "立即运行",
        edit: "编辑",
        delete: "删除",
        confirm_delete: "确定要删除吗？",
        no_tasks: "未发现任务。点击“添加新任务”开始。",
        running: "运行中",
        
        // Sections
        basic_info: "基本信息",
        connection: "Openlist 连接",
        paths: "路径设置",
        advanced_settings: "高级设置与功能",
        smart_protection: "智能保护",

        // Fields
        task_id: "任务 ID",
        task_id_help: "标识 ID",
        cron: "Cron 表达式",
        cron_help: "后台定时任务 Cron 表达式",
        url: "服务器地址",
        url_help: "Openlist 服务器地址",
        public_url: "公共地址",
        public_url_help: "用于 OpenlistURL 模式生成内容的地址 (可选)",
        username: "用户名",
        username_help: "Openlist 用户名",
        password: "密码",
        password_help: "Openlist 密码",
        token: "令牌 (Token)",
        token_help: "永久令牌 (可选，填入后无需账号密码)",
        source_dir: "源码目录",
        source_dir_help: "服务器上文件夹路径",
        target_dir: "目标目录",
        target_dir_help: "本地输出路径",
        next_run: "下一次运行",
        last_error: "最后一次错误",
        flatten_mode: "平铺模式",
        flatten_mode_help: "开启后 subtitle、image、nfo 强制关闭",
        subtitle: "下载字幕",
        subtitle_help: "是否下载字幕文件",
        image: "下载图片",
        image_help: "是否下载图片文件",
        nfo: "下载 NFO",
        nfo_help: "是否下载 .nfo 文件",
        mode: "模式",
        mode_help: "Strm 文件中的内容模式",
        overwrite: "覆盖模式",
        overwrite_help: "本地存在同名文件时是否重新生成",
        sync_server: "同步服务器",
        sync_server_help: "是否与服务器状态同步",
        sync_ignore: "同步忽略正则",
        sync_ignore_help: "同步时忽略的文件正则表达式",
        sync_back: "同步回服务器",
        sync_back_help: "将本地删除同步回远程服务器 (慎用)",
        other_ext: "自定义后缀",
        other_ext_help: "自定义下载后缀，半角逗号分隔",
        check_mod_time: "目录时间校验",
        check_mod_time_help: "如果目录修改时间未变，则跳过扫描 (注意：部分网盘可能无法识别删除)",
        max_workers: "最大并发数",
        max_workers_help: "减轻对 Openlist 服务器的负载",
        max_downloaders: "最大下载数",
        max_downloaders_help: "最大同时下载文件数",
        wait_time: "间隔时间 (秒)",
        wait_time_help: "遍历请求间隔时间，避免风控",
        enabled: "启用",
        enabled_help: "防止 Openlist 故障导致的大量删除",
        threshold: "阈值",
        threshold_help: "触发保护的文件数量阈值",
        grace_scans: "扫描次数",
        grace_scans_help: "删除前需要的扫描次数",

        // Login
        login_title: "StrmHelper 登录",
        login_btn: "登录",
        invalid_creds: "凭据无效"
    }
};

let currentLang = localStorage.getItem('lang') || 'en';

function i18n(key) {
    return translations[currentLang][key] || key;
}

function applyTranslations() {
    document.querySelectorAll('[data-i18n]').forEach(el => {
        const key = el.getAttribute('data-i18n');
        if (el.tagName === 'INPUT' && (el.type === 'text' || el.type === 'password' || el.type === 'number')) {
            el.placeholder = i18n(key);
        } else if (el.tagName === 'OPTGROUP') {
            el.label = i18n(key);
        } else {
            el.innerText = i18n(key);
        }
    });

    document.querySelectorAll('[data-i18n-placeholder]').forEach(el => {
        const key = el.getAttribute('data-i18n-placeholder');
        el.placeholder = i18n(key);
    });

    document.querySelectorAll('[data-i18n-title]').forEach(el => {
        const key = el.getAttribute('data-i18n-title');
        el.title = i18n(key);
    });
    
    document.documentElement.lang = currentLang;
}

function switchLang(lang) {
    currentLang = lang;
    localStorage.setItem('lang', lang);
    applyTranslations();
}

// Observe DOM changes to apply translations to new elements (like HTMX swaps)
const observer = new MutationObserver((mutations) => {
    // Check if any added nodes actually need translation to avoid unnecessary triggers
    const hasNewTranslatable = Array.from(mutations).some(m => 
        Array.from(m.addedNodes).some(node => 
            node.nodeType === 1 && (node.hasAttribute('data-i18n') || node.querySelector('[data-i18n]'))
        )
    );
    
    if (hasNewTranslatable) {
        observer.disconnect();
        applyTranslations();
        observer.observe(document.body, { childList: true, subtree: true });
    }
});

document.addEventListener('DOMContentLoaded', () => {
    applyTranslations();
    observer.observe(document.body, { childList: true, subtree: true });
});
