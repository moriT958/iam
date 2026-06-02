---
title: "JavaScript の非同期処理の基礎を学んだ"
date: "2025-08-09"
tags: ["JavaScript", "TypeScript"]
description: "JavaScript における非同期処理まとめ"
external: false
draft: false
---

JavaScript における非同期処理について学んだことを備忘録としてまとめました。

## Javascript のスレッド

JavaScript はシングルスレッド上で実行されます。
ブラウザには主に以下の３つのスレッドが存在し、JavaScript が実行されるのは Main Thread です。

- Main Thread (JS の実行とレンダリング処理を行う)
- Service Worker
- Web Worker

## 同期処理と非同期処理

- 同期処理: メインスレッド上で順番に処理が進んでいきます
- 非同期処理: 一時的にメインスレッドから処理が切り離されます

例) setTimeout(callback, ms)

setTimeout を実行した時点でメインスレッドから切り離されます。
そして非同期 API (内部で実装されたタイマー) に渡されたのち、指定した時間後にコールバックがタスクキューに積まれます。

## タスクキューとコールスタック

重要な登場人物一覧:

1. コールスタック

- 実行中のコンテキストが積まれ、処理が終わると取り出されます
- コンテキストには、関数コンテキストやグローバルコンテキストがあります
- コンテキストが持つ情報:
  - ローカル変数や引数の値
  - 計算途中の値
  - 関数終了後にどこに戻るか

2. イベントループ

- コールスタックにコンテキストが積まれているかを監視しています
- コールスタックが空になったタイミングで、タスクキューから次の処理を取り出します

3. タスクキュー

- 実行待ちの関数行列 (FIFO)
- コールスタックが空の場合、積まれている関数を実行します
- 関数の実行結果はコンテキストとしてコールスタックに積まれます
- 1 つのループあたり実行できる関数は 1 個まです

## Promise

| 状態        | 意味                 | 遷移条件                 |
| ----------- | -------------------- | ------------------------ |
| `pending`   | 初期状態 (処理中)    | Promise が生成された直後 |
| `fulfilled` | 成功して完了した状態 | `resolve()` が呼ばれた時 |
| `rejected`  | 失敗した状態         | `reject()` が呼ばれた時  |

Promise のコード例

```javascript
const myPromise = new Promise(function (resolve, reject) {
  const success = true;

  if (success) {
    resolve("Success!"); // fulfilled 状態にする
  } else {
    reject(new Error("Failed..")); // rejected 状態にする
  }
});

myPromise
  .then(function (result) {
    console.log(result); // -> "Success!"
    return result + " (done)"; // 次の then に値を渡す
  })
  .then(function (result) {
    console.log(result); // -> "Success! (done)"
  })
  .catch(function (error) {
    // reject の引数、または throw されたエラーが渡ってくる
    console.error(error);
  })
  .finally(function () {
    // 成功・失敗に関わらず必ず実行される
    console.log("done");
  });
```

## MicroTasks と MacroTasks

キューには２種類あります。

### Macro Tasks (タスクキュー)

- setTimeout や setInterval のコールバックなどの非同期処理が積まれます
- 1 回のイベントループあたりに 1 つのタスクを実行します

### Micro Tasks (ジョブキュー)

- Promise の then などの非同期処理が積まれます
- タスクキューよりもジョブキューの方が先に実行されます
- 1 回のイベントループあたりに積まれている全てのジョブが実行されます

### コードの実行順序

以下のコードを実行した時の実行順はどうなるでしょうか？

```javascript
console.log("starting");

new Promise(function (resolve) {
  setTimeout(function () {
    console.log("task1");
  });
  resolve();
})
  .then(function () {
    console.log("job1");
  })
  .then(function () {
    console.log("job2");
  });

console.log("Global context end");
```

答え:

```txt
starting
global context end
job1
job2
task1
```

処理の流れ:

1. コールスタックにグローバルコンテキストが積まれます
1. "starting" が表示されます
1. setTimeout が呼ばれ、タイマーが開始されます
1. resolve() が呼ばれ、"job1" のハンドラがジョブキューに登録されます
1. "global context end" が表示されます (グローバルコンテキストがコールスタックから取り出される)
1. コールスタックが空になり、ジョブキューの "job1" が実行されます
1. "job1" が表示されます。このタイミングで "job2" のハンドラがジョブキューに登録されます
1. "job2" が実行されます
1. "job2" が表示されます
1. ジョブキューが空になり、タスクキューの "task1" が実行されます
1. "task1" が表示されます

## await/async

- async で定義されている関数の返り値は必ず Promise になります
- await は、Promise を返す関数の非同期処理が完了するのを待ちます

コード例

```javascript
const fetchCoffee = async () => {
  const response = await fetch("https://api.sampleapis.com/coffee/hot");
  const coffeeList = await response.json();
  console.log(coffeeList);
};

fetchCoffee();
```

## Promise の静的メソッドを使った並行処理

Promise の並行処理に関する静的メソッド一覧です。

### 1. Promise.all

引数には Promise を格納した反復可能オブジェクトを入れます。
全ての Promise が `fulfilled` となった場合のみ成功扱いになります。
1 つでも `rejected` となると、失敗となります。

返り値は、以下のような Promise を 1 つ返します。

```txt
// 成功時
Promise { <state>: "fulfilled", <value>: Array[5] }

// 失敗時
Promise { <state>: "rejected", <reason>: 5 }
```

### 2. Promise.race

最も早く `pending` 状態が終わった Promise を返します。
そのため、返ってくる Promise は `fulfilled` か `rejected` のどちらかわかりません。
また、どの Promise も解決してない場合は `pending` となります。

返り値の例:

```txt
// 成功時
Promise { status: 'fulfilled', value: 100 }

// 失敗時
Promise { status: 'rejected', reason: 300 }

// 未解決時
Promise { status: 'pending' }
```

### 3. Promise.allSettled

全ての Promise が解決する (`pending` 状態から抜ける) とそれら全てを返します。

返り値の例:

```txt
[
   { status: 'fulfilled', value: 33 },
   { status: 'fulfilled', value: 66 },
   { status: 'rejected', reason: Error: an error }
]
```

### Promise.all を使用した並行処理の実例

リストの要素を並行処理でインクリメントします。

```javascript
const inc = (n) =>
  new Promise((resolve) => setTimeout(() => resolve(n + 1), 100));

const incNums = async (nums) => {
  return Promise.all(nums.map((n) => inc(n)))
    .then((result) => {
      if (result.length === 0) {
        throw new Error("failed to increment all nums");
      }
      return result;
    })
    .catch((err) => {
      console.error(err);
      throw err;
    });
};

const nums = [1, 2, 3, 4, 5];
console.log(nums);
incNums(nums).then((result) => {
  console.log(result);
});
```

```bash
# ==== 出力 ====
[ 1, 2, 3, 4, 5 ]
[ 2, 3, 4, 5, 6 ]
```

## setTimeout を応用した定期実行

setTimeout を使用することで定期実行を行う処理を実現することができます。

setInterval ではなく、setTimeout を使うことで安全に定期実行を行うことができます。

setInterval は一定時間間隔でタスクキューにタスクを積むため、前のタスクに時間がかかってしまっても、
止めない限りお構いなしにタスクを積んでいってしまいます。そのためキューが詰まってしまうバグに繋がります。

一方で、setTimeout を使用した場合、タスクの実行を待ってから次のタスクを積むといった具合に、
柔軟なスケジューリングが実装できます。

以下は、指定した回数分コンソールに数字を出力する処理を setTimeout で実装した例です。
コメントを記している箇所で万が一時間のかかる処理があったとしても、その実行を待ってからタスクを積むため、
キューが詰まるリスクを回避することができます。

```javascript
const count = (num) => {
  let count = 0;
  let timer;

  const timerFunc = () => {
    if (count >= num) {
      clearTimeout(timer);
      return;
    }

    // 時間がかかる処理

    count++;
    console.log(count);
    timer = setTimeout(timerFunc, 1000);
  };

  timer = setTimeout(timerFunc, 1000);
};
```

## まとめ

- JavaScript はシングルスレッドである
- 非同期処理はコールスタック・イベントループ・タスクキューの連携で動作する
- Promise の状態は `pending` / `fulfilled` / `rejected` の 3 種類がある
- Microtask (Promise.then) は Macrotask (setTimeout) より先に実行される
- `Promise.all` / `Promise.race` / `Promise.allSettled` で並行処理が書ける
- 定期実行は `setInterval` より `setTimeout` の再帰呼び出しが安全

## おまけ

JS ランタイムの構成要素たち (色々調べてて面白かったのでおまけに)

| 要素            | 場所      | 役割                                                      |
| --------------- | --------- | --------------------------------------------------------- |
| Heap            | JS Engine | オブジェクトのメモリ割り当て                              |
| Call Stack      | JS Engine | 実行コンテキストの管理 (LIFO)                             |
| Web APIs        | ブラウザ  | setTimeout / fetch / DOM events など非同期 API の実装     |
| Microtask Queue | Runtime   | Promise.then / queueMicrotask のコールバック待ち行列      |
| Macrotask Queue | Runtime   | setTimeout / setInterval / I/O のコールバック待ち行列     |
| Event Loop      | Runtime   | Stack が空になったら Queue からタスクを Call Stack へ送る |

<figure>
  <img src="https://miro.medium.com/1*0YlG8zk5rRCEbkndh8_Pfw.jpeg" alt="JS ランタイム構成図" />
  <figcaption>出典: <a href="https://medium.com/@el.elhamhashemi/understanding-js-runtime-js-engine-execution-context-a6f37c5d7f8f" target="_blank" rel="noopener noreferrer">Understanding JS Runtime, JS Engine & Execution Context - Elham Hashemi</a></figcaption>
</figure>

**具体的な実装例**

| 要素            | ブラウザ                           | Node.js     | Deno                  | Bun              |
| --------------- | ---------------------------------- | ----------- | --------------------- | ---------------- |
| JS Engine       | V8 / SpiderMonkey / JavaScriptCore | V8          | V8                    | JavaScriptCore   |
| 非同期 I/O      | ブラウザ内蔵 Web APIs              | libuv       | Tokio                 | 独自実装 (C/Zig) |
| Event Loop      | ブラウザ実装                       | libuv       | Tokio + Deno 独自実装 | 独自実装         |
| Microtask Queue | ブラウザ実装                       | V8 組み込み | V8 組み込み           | JSC 組み込み     |

- **V8**: Google 製の JS エンジン。Heap と Call Stack を担当し、JIT コンパイルで高速実行を実現。Chrome / Edge / Node.js / Deno で採用されている。
- **SpiderMonkey**: Mozilla 製の JS エンジン。Firefox で使用されており、V8 と同様に JIT コンパイルをサポートする。
- **JavaScriptCore** (JSC): Apple 製の JS エンジン。Safari / iOS の WKWebView で使用される。Bun もこれを採用しており、V8 より起動速度が速いとされる。
- **libuv**: Node.js のために開発された C ライブラリ。非同期 I/O・タイマー・イベントループを提供し、OS ごとの差異 (Linux: epoll / macOS: kqueue / Windows: IOCP) を抽象化している。
- **Tokio**: Deno が採用する Rust 製の非同期ランタイム。libuv と同様の役割を担う。
- **Bun**: Zig で書かれた高速な JS ランタイム。JSC を JS エンジンとして採用し、非同期 I/O やイベントループも独自実装することで Node.js より高いスループットを実現している。
