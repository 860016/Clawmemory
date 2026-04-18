<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <h1 class="text-2xl font-bold text-gray-900 mb-6">仪表盘</h1>

    <!-- 安全警报 -->
    <?php if (!empty($recentFailedLogins)): ?>
    <div class="mb-6 bg-red-50 border border-red-200 rounded-lg p-4">
        <h3 class="text-sm font-semibold text-red-800 mb-2">⚠ 登录安全警报（近24小时）</h3>
        <div class="space-y-1">
            <?php foreach ($recentFailedLogins as $log): ?>
            <p class="text-xs text-red-700">
                <?= date('H:i', strtotime($log['created_at'])) ?> - 
                <?= htmlspecialchars($log['email']) ?> 来自 
                <code class="bg-red-100 px-1 rounded"><?= htmlspecialchars($log['ip_address']) ?></code> 登录失败
            </p>
            <?php endforeach; ?>
        </div>
        <a href="/admin/settings" onclick="showTab('security')" class="text-xs text-red-600 hover:text-red-800 mt-2 inline-block">查看全部 →</a>
    </div>
    <?php endif; ?>

    <!-- 即将到期 -->
    <?php if (!empty($expiringSoon)): ?>
    <div class="mb-6 bg-yellow-50 border border-yellow-200 rounded-lg p-4">
        <h3 class="text-sm font-semibold text-yellow-800 mb-2">⏰ 即将到期的授权码（7天内）</h3>
        <div class="space-y-1">
            <?php foreach ($expiringSoon as $lic): ?>
            <p class="text-xs text-yellow-700">
                <code class="bg-yellow-100 px-1 rounded"><?= htmlspecialchars($lic['license_key']) ?></code>
                (<?= strtoupper($lic['tier']) ?>) - 
                <?= htmlspecialchars($lic['customer_name'] ?? '未命名') ?> -
                到期：<?= date('Y-m-d', strtotime($lic['expires_at'])) ?>
                <?php if (!empty($lic['customer_email'])): ?>
                📧 <?= htmlspecialchars($lic['customer_email']) ?>
                <?php endif; ?>
            </p>
            <?php endforeach; ?>
        </div>
    </div>
    <?php endif; ?>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <div class="bg-white rounded-lg shadow p-6">
            <div class="text-sm font-medium text-gray-500">总授权码</div>
            <div class="mt-1 text-3xl font-bold text-gray-900"><?= $totalLicenses ?></div>
        </div>
        <div class="bg-white rounded-lg shadow p-6">
            <div class="text-sm font-medium text-gray-500">活跃授权</div>
            <div class="mt-1 text-3xl font-bold text-green-600"><?= $activeLicenses ?></div>
        </div>
        <div class="bg-white rounded-lg shadow p-6">
            <div class="text-sm font-medium text-gray-500">绑定设备</div>
            <div class="mt-1 text-3xl font-bold text-blue-600"><?= $totalDevices ?></div>
        </div>
        <div class="bg-white rounded-lg shadow p-6">
            <div class="text-sm font-medium text-gray-500">今日活跃</div>
            <div class="mt-1 text-3xl font-bold text-purple-600"><?= $todayActive ?></div>
        </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <div class="bg-white rounded-lg shadow p-6">
            <div class="text-sm font-medium text-gray-500">已过期</div>
            <div class="mt-1 text-2xl font-bold text-orange-500"><?= $expiredLicenses ?></div>
        </div>
        <div class="bg-white rounded-lg shadow p-6">
            <div class="text-sm font-medium text-gray-500">已吊销</div>
            <div class="mt-1 text-2xl font-bold text-red-500"><?= $revokedLicenses ?></div>
        </div>
        <div class="bg-white rounded-lg shadow p-6">
            <div class="text-sm font-medium text-gray-500">总安装量</div>
            <div class="mt-1 text-2xl font-bold text-gray-900"><?= $totalInstalls ?></div>
        </div>
        <div class="bg-white rounded-lg shadow p-6">
            <div class="text-sm font-medium text-gray-500">月度收入</div>
            <div class="mt-1 text-2xl font-bold text-green-600">$<?= number_format((float) $monthlyRevenue, 2) ?></div>
        </div>
    </div>

    <!-- 快捷操作 -->
    <div class="bg-white rounded-lg shadow p-6">
        <h2 class="text-lg font-semibold text-gray-900 mb-4">快捷操作</h2>
        <div class="flex flex-wrap gap-3">
            <a href="/admin/licenses/create" class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700">创建授权码</a>
            <a href="/admin/licenses" class="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md shadow-sm text-gray-700 bg-white hover:bg-gray-50">查看授权码</a>
            <a href="/admin/devices" class="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md shadow-sm text-gray-700 bg-white hover:bg-gray-50">设备管理</a>
            <a href="/admin/settings" class="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md shadow-sm text-gray-700 bg-white hover:bg-gray-50">系统设置</a>
        </div>
    </div>
</div>
