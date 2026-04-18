<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <h1 class="text-2xl font-bold text-gray-900 mb-6">使用统计</h1>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- DAU 趋势 -->
        <div class="bg-white shadow sm:rounded-lg p-6">
            <h2 class="text-lg font-semibold text-gray-900 mb-4">DAU 趋势（近30天）</h2>
            <?php if (empty($dauTrend)): ?>
            <p class="text-gray-500">暂无数据</p>
            <?php else: ?>
            <div class="space-y-2">
                <?php 
                $maxDau = max(array_column($dauTrend, 'dau')) ?: 1;
                foreach ($dauTrend as $row): 
                    $percent = ($row['dau'] / $maxDau) * 100;
                ?>
                <div class="flex items-center space-x-3">
                    <span class="text-xs text-gray-500 w-20"><?= $row['stat_date'] ?></span>
                    <div class="flex-1 bg-gray-200 rounded-full h-4">
                        <div class="bg-indigo-600 h-4 rounded-full" style="width: <?= $percent ?>%"></div>
                    </div>
                    <span class="text-sm font-medium text-gray-700 w-8"><?= $row['dau'] ?></span>
                </div>
                <?php endforeach; ?>
            </div>
            <?php endif; ?>
        </div>

        <!-- 版本分布 -->
        <div class="bg-white shadow sm:rounded-lg p-6">
            <h2 class="text-lg font-semibold text-gray-900 mb-4">版本分布</h2>
            <?php if (empty($versionDist)): ?>
            <p class="text-gray-500">暂无数据</p>
            <?php else: ?>
            <div class="space-y-3">
                <?php foreach ($versionDist as $row): ?>
                <div class="flex justify-between items-center">
                    <span class="text-sm text-gray-700"><?= htmlspecialchars($row['app_version'] ?? '未知') ?></span>
                    <span class="text-sm font-medium text-gray-900"><?= $row['count'] ?> 设备</span>
                </div>
                <?php endforeach; ?>
            </div>
            <?php endif; ?>
        </div>

        <!-- 套餐分布 -->
        <div class="bg-white shadow sm:rounded-lg p-6">
            <h2 class="text-lg font-semibold text-gray-900 mb-4">套餐分布</h2>
            <?php if (empty($tierDist)): ?>
            <p class="text-gray-500">暂无数据</p>
            <?php else: ?>
            <div class="space-y-3">
                <?php foreach ($tierDist as $row): ?>
                <div class="flex justify-between items-center">
                    <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full 
                        <?= $row['tier'] === 'enterprise' ? 'bg-purple-100 text-purple-800' : ($row['tier'] === 'pro' ? 'bg-blue-100 text-blue-800' : 'bg-gray-100 text-gray-800') ?>">
                        <?= strtoupper($row['tier']) ?>
                    </span>
                    <span class="text-sm font-medium text-gray-900"><?= $row['count'] ?> 个</span>
                </div>
                <?php endforeach; ?>
            </div>
            <?php endif; ?>
        </div>
    </div>
</div>
