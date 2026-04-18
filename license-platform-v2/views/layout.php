<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title><?= htmlspecialchars($config['app_name'] ?? 'OpenClaw License') ?></title>
    <script src="https://cdn.tailwindcss.com"></script>
    <style>
        [x-cloak] { display: none !important; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; }
    </style>
</head>
<body class="bg-gray-50 min-h-screen">

<?php if ($isLoggedIn && basename(($_SERVER['REQUEST_URI'] ?? '/'), '.php') !== 'login'): ?>
<!-- 顶部导航 -->
<nav class="bg-white shadow-sm border-b border-gray-200">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16">
            <div class="flex items-center">
                <a href="/admin/dashboard" class="text-xl font-bold text-indigo-600">OpenClaw License</a>
                <div class="hidden md:flex ml-10 space-x-4">
                    <a href="/admin/dashboard" class="text-gray-600 hover:text-indigo-600 px-3 py-2 rounded-md text-sm font-medium">仪表盘</a>
                    <a href="/admin/licenses" class="text-gray-600 hover:text-indigo-600 px-3 py-2 rounded-md text-sm font-medium">授权码</a>
                    <a href="/admin/devices" class="text-gray-600 hover:text-indigo-600 px-3 py-2 rounded-md text-sm font-medium">设备管理</a>
                    <a href="/admin/stats" class="text-gray-600 hover:text-indigo-600 px-3 py-2 rounded-md text-sm font-medium">统计</a>
                    <a href="/admin/pricing" class="text-gray-600 hover:text-indigo-600 px-3 py-2 rounded-md text-sm font-medium">价格方案</a>
                </div>
            </div>
            <div class="flex items-center space-x-4">
                <a href="/admin/settings" class="text-gray-500 hover:text-indigo-600 text-sm" title="系统设置">
                    <svg class="w-5 h-5 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"></path><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path></svg>
                    设置
                </a>
                <span class="text-sm text-gray-500"><?= htmlspecialchars($currentUserEmail) ?></span>
                <a href="/admin/change-password" class="text-gray-500 hover:text-indigo-600 text-sm">改密码</a>
                <form method="POST" action="/logout" class="inline">
                    <button type="submit" class="text-gray-500 hover:text-red-600 text-sm">退出</button>
                </form>
            </div>
        </div>
    </div>
</nav>
<?php endif; ?>

<main>
    <?= $content ?>
</main>

</body>
</html>
