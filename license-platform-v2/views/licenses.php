<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <div class="flex justify-between items-center mb-6">
        <h1 class="text-2xl font-bold text-gray-900">授权码管理</h1>
        <a href="/admin/licenses/create" class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700">
            创建授权码
        </a>
    </div>

    <?php if (isset($_GET['created'])): ?>
    <div class="mb-4 bg-green-50 border border-green-200 text-green-700 px-4 py-3 rounded">
        授权码创建成功：<code class="bg-green-100 px-2 py-1 rounded font-mono"><?= htmlspecialchars($_GET['created']) ?></code>
    </div>
    <?php endif; ?>

    <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
                <tr>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">授权码</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">套餐</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">状态</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">设备</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">客户</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">过期时间</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">操作</th>
                </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
                <?php foreach ($licenses as $license): ?>
                <tr>
                    <td class="px-6 py-4 whitespace-nowrap">
                        <code class="text-sm font-mono bg-gray-100 px-2 py-1 rounded"><?= htmlspecialchars($license['license_key']) ?></code>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap">
                        <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full 
                            <?= $license['tier'] === 'enterprise' ? 'bg-purple-100 text-purple-800' : ($license['tier'] === 'pro' ? 'bg-blue-100 text-blue-800' : 'bg-gray-100 text-gray-800') ?>">
                            <?= strtoupper($license['tier']) ?>
                        </span>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap">
                        <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full 
                            <?= $license['status'] === 'active' ? 'bg-green-100 text-green-800' : ($license['status'] === 'expired' ? 'bg-yellow-100 text-yellow-800' : 'bg-red-100 text-red-800') ?>">
                            <?= $license['status'] === 'active' ? '活跃' : ($license['status'] === 'expired' ? '已过期' : '已吊销') ?>
                        </span>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        <?= $license['device_count'] ?>/<?= $license['max_devices'] ?>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        <?= htmlspecialchars($license['customer_name'] ?? '-') ?>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        <?= $license['expires_at'] ? date('Y-m-d', strtotime($license['expires_at'])) : '永久' ?>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm space-x-2">
                        <?php if ($license['status'] === 'active'): ?>
                        <form method="POST" action="/admin/licenses/revoke" class="inline">
                            <input type="hidden" name="id" value="<?= $license['id'] ?>">
                            <button type="submit" class="text-red-600 hover:text-red-900" onclick="return confirm('确定要吊销此授权码吗？')">吊销</button>
                        </form>
                        <?php endif; ?>
                        <?php if ($license['status'] !== 'active'): ?>
                        <form method="POST" action="/admin/licenses/renew" class="inline">
                            <input type="hidden" name="id" value="<?= $license['id'] ?>">
                            <button type="submit" class="text-green-600 hover:text-green-900">续期</button>
                        </form>
                        <?php endif; ?>
                    </td>
                </tr>
                <?php endforeach; ?>
                <?php if (empty($licenses)): ?>
                <tr>
                    <td colspan="7" class="px-6 py-4 text-center text-gray-500">暂无授权码</td>
                </tr>
                <?php endif; ?>
            </tbody>
        </table>
    </div>

    <!-- 分页 -->
    <?php if ($totalPages > 1): ?>
    <div class="mt-4 flex justify-center space-x-2">
        <?php for ($i = 1; $i <= $totalPages; $i++): ?>
        <a href="?page=<?= $i ?>" class="px-3 py-1 rounded <?= $i === $page ? 'bg-indigo-600 text-white' : 'bg-white text-gray-700 border' ?>">
            <?= $i ?>
        </a>
        <?php endfor; ?>
    </div>
    <?php endif; ?>
</div>
