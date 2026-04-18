<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <div class="flex justify-between items-center mb-6">
        <h1 class="text-2xl font-bold text-gray-900">设备管理</h1>
    </div>

    <!-- 搜索 -->
    <div class="mb-6">
        <form method="GET" action="/admin/devices" class="flex space-x-2">
            <input type="text" name="search" value="<?= htmlspecialchars($search) ?>" placeholder="搜索设备指纹或名称..."
                class="flex-1 rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm border p-2">
            <button type="submit" class="px-4 py-2 bg-indigo-600 text-white text-sm rounded-md hover:bg-indigo-700">搜索</button>
            <?php if ($search): ?>
            <a href="/admin/devices" class="px-4 py-2 border border-gray-300 text-sm rounded-md text-gray-700 hover:bg-gray-50">清除</a>
            <?php endif; ?>
        </form>
    </div>

    <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
                <tr>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">设备名称</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">设备指纹</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">授权码</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">套餐</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">系统</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">最后活跃</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">操作</th>
                </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
                <?php foreach ($devices as $device): ?>
                <tr>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        <?= htmlspecialchars($device['device_name']) ?>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap">
                        <code class="text-xs font-mono bg-gray-100 px-2 py-1 rounded"><?= htmlspecialchars(substr($device['fingerprint_hash'], 0, 16)) ?>...</code>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap">
                        <code class="text-xs font-mono bg-gray-100 px-2 py-1 rounded"><?= htmlspecialchars($device['license_key']) ?></code>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap">
                        <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full 
                            <?= $device['tier'] === 'enterprise' ? 'bg-purple-100 text-purple-800' : ($device['tier'] === 'pro' ? 'bg-blue-100 text-blue-800' : 'bg-gray-100 text-gray-800') ?>">
                            <?= strtoupper($device['tier']) ?>
                        </span>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        <?= htmlspecialchars($device['os_info'] ?? '-') ?>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        <?= $device['last_active_at'] ? date('Y-m-d H:i', strtotime($device['last_active_at'])) : '-' ?>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm">
                        <form method="POST" action="/admin/devices/delete" class="inline">
                            <input type="hidden" name="id" value="<?= $device['id'] ?>">
                            <button type="submit" class="text-red-600 hover:text-red-900" onclick="return confirm('确定要移除此设备吗？')">移除</button>
                        </form>
                    </td>
                </tr>
                <?php endforeach; ?>
                <?php if (empty($devices)): ?>
                <tr>
                    <td colspan="7" class="px-6 py-4 text-center text-gray-500">暂无设备</td>
                </tr>
                <?php endif; ?>
            </tbody>
        </table>
    </div>

    <!-- 分页 -->
    <?php if ($totalPages > 1): ?>
    <div class="mt-4 flex justify-center space-x-2">
        <?php for ($i = 1; $i <= $totalPages; $i++): ?>
        <a href="?page=<?= $i ?><?= $search ? '&search=' . urlencode($search) : '' ?>" 
           class="px-3 py-1 rounded <?= $i === $page ? 'bg-indigo-600 text-white' : 'bg-white text-gray-700 border' ?>">
            <?= $i ?>
        </a>
        <?php endfor; ?>
    </div>
    <?php endif; ?>
</div>
