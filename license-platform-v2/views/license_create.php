<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <div class="flex justify-between items-center mb-6">
        <h1 class="text-2xl font-bold text-gray-900">创建授权码</h1>
        <a href="/admin/licenses" class="text-gray-600 hover:text-gray-900 text-sm">返回列表</a>
    </div>

    <?php if (!empty($error)): ?>
    <div class="mb-4 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
        <?= htmlspecialchars($error) ?>
    </div>
    <?php endif; ?>

    <div class="bg-white shadow sm:rounded-lg p-6">
        <form method="POST" action="/admin/licenses/create" class="space-y-6">
            <!-- 套餐选择 -->
            <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">套餐类型</label>
                <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <?php foreach ($tiers as $key => $tier): ?>
                    <label class="relative flex cursor-pointer rounded-lg border p-4 shadow-sm focus-within:ring-2 focus-within:ring-indigo-500 
                        <?= ($key === 'pro') ? 'border-indigo-500 ring-2 ring-indigo-500' : 'border-gray-300' ?>">
                        <input type="radio" name="tier" value="<?= $key ?>" <?= ($key === 'pro') ? 'checked' : '' ?>
                            class="mt-1 h-4 w-4 border-gray-300 text-indigo-600 focus:ring-indigo-500"
                            onchange="this.closest('.grid').querySelectorAll('label').forEach(l=>l.classList.remove('border-indigo-500','ring-2','ring-indigo-500'));this.closest('label').classList.add('border-indigo-500','ring-2','ring-indigo-500');document.getElementById('max_devices').value=<?= $tier['max_devices'] ?>">
                        <div class="ml-3 flex-1">
                            <span class="block text-sm font-medium text-gray-900"><?= $tier['name'] ?></span>
                            <span class="block text-sm text-gray-500">$<?= $tier['price'] ?>/月 · 最多 <?= $tier['max_devices'] ?> 设备</span>
                            <?php if (!empty($tier['features'])): ?>
                            <span class="block text-xs text-gray-400 mt-1"><?= implode(', ', $tier['features']) ?></span>
                            <?php endif; ?>
                        </div>
                    </label>
                    <?php endforeach; ?>
                </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                    <label for="max_devices" class="block text-sm font-medium text-gray-700">最大设备数</label>
                    <input type="number" id="max_devices" name="max_devices" value="3" min="1"
                        class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm border p-2">
                </div>
                <div>
                    <label for="expires_at" class="block text-sm font-medium text-gray-700">过期时间（留空为永久）</label>
                    <input type="date" id="expires_at" name="expires_at"
                        class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm border p-2">
                </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                    <label for="customer_name" class="block text-sm font-medium text-gray-700">客户名称</label>
                    <input type="text" id="customer_name" name="customer_name"
                        class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm border p-2">
                </div>
                <div>
                    <label for="customer_email" class="block text-sm font-medium text-gray-700">客户邮箱</label>
                    <input type="email" id="customer_email" name="customer_email"
                        class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm border p-2">
                </div>
            </div>

            <div>
                <label for="notes" class="block text-sm font-medium text-gray-700">备注</label>
                <textarea id="notes" name="notes" rows="3"
                    class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm border p-2"></textarea>
            </div>

            <div>
                <button type="submit"
                    class="inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">
                    创建授权码
                </button>
            </div>
        </form>
    </div>
</div>
