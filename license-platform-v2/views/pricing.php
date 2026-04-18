<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <h1 class="text-2xl font-bold text-gray-900 mb-6">价格方案</h1>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
        <?php 
        $config = getConfig();
        $labels = $config['feature_labels'] ?? [];
        foreach ($tiers as $key => $tier): ?>
        <div class="bg-white rounded-2xl shadow-lg overflow-hidden 
            <?= $key === 'pro' ? 'ring-2 ring-indigo-500' : '' ?>">
            <?php if ($key === 'pro'): ?>
            <div class="bg-indigo-600 text-white text-center py-2 text-sm font-medium">最受欢迎</div>
            <?php endif; ?>
            <div class="p-8">
                <h3 class="text-2xl font-bold text-gray-900"><?= htmlspecialchars($tier['name'] ?? ucfirst($key)) ?></h3>
                <div class="mt-4">
                    <span class="text-4xl font-extrabold text-gray-900">$<?= $tier['price'] ?></span>
                    <span class="text-gray-500">/月</span>
                </div>
                <p class="mt-2 text-sm text-gray-500">最多 <?= $tier['max_devices'] ?> 台设备</p>

                <ul class="mt-6 space-y-3">
                    <?php if (empty($tier['features'])): ?>
                    <li class="flex items-start">
                        <span class="text-green-500 mr-2">✓</span>
                        <span class="text-sm text-gray-700">基础记忆 CRUD</span>
                    </li>
                    <li class="flex items-start">
                        <span class="text-green-500 mr-2">✓</span>
                        <span class="text-sm text-gray-700">FTS5 全文搜索</span>
                    </li>
                    <li class="flex items-start">
                        <span class="text-green-500 mr-2">✓</span>
                        <span class="text-sm text-gray-700">知识库 (Knowledge)</span>
                    </li>
                    <li class="flex items-start">
                        <span class="text-green-500 mr-2">✓</span>
                        <span class="text-sm text-gray-700">单用户 / 单设备</span>
                    </li>
                    <?php else: ?>
                    <?php foreach ($tier['features'] as $feature): ?>
                    <li class="flex items-start">
                        <span class="text-green-500 mr-2">✓</span>
                        <span class="text-sm text-gray-700"><?= htmlspecialchars($labels[$feature] ?? $feature) ?></span>
                    </li>
                    <?php endforeach; ?>
                    <?php endif; ?>
                </ul>
            </div>
        </div>
        <?php endforeach; ?>
    </div>
</div>
