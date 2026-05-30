# FV-SEC001 - Thử thách dành cho Software Engineer — Ad Performance Aggregator

## Giới thiệu

Đây là một bài test xử lý dữ liệu dành cho ứng viên Software Engineer ứng tuyển vào công ty chúng tôi.

Bạn sẽ làm việc với một bộ dữ liệu CSV lớn (~1GB) chứa các bản ghi hiệu suất quảng cáo.

Mục tiêu là đánh giá khả năng của bạn trong việc:

- viết code sạch
- xử lý dữ liệu lớn hiệu quả
- tối ưu hiệu năng và bộ nhớ
- thiết kế workflow xử lý dữ liệu ổn định và đáng tin cậy

---

# Dữ liệu đầu vào

## Tải dataset

1. Tải file `ad_data.csv.zip` từ thư mục trong repository này
2. Giải nén để lấy file `ad_data.csv` (~1GB)
3. Sử dụng file CSV này cho bài làm của bạn

```bash
# Ví dụ: giải nén file
unzip ad_data.csv.zip
```

## Schema của CSV

| Column      | Type    | Description                   |
| ----------- | ------- | ----------------------------- |
| campaign_id | string  | ID chiến dịch                 |
| date        | string  | Ngày theo format `YYYY-MM-DD` |
| impressions | integer | Số lượt hiển thị              |
| clicks      | integer | Số lượt click                 |
| spend       | float   | Chi phí quảng cáo (USD)       |
| conversions | integer | Số lượt chuyển đổi            |

## Ví dụ

| campaign_id | date       | impressions | clicks | spend | conversions |
| ----------- | ---------- | ----------- | ------ | ----- | ----------- |
| CMP001      | 2025-01-01 | 12000       | 300    | 45.50 | 12          |
| CMP002      | 2025-01-01 | 8000        | 120    | 28.00 | 4           |
| CMP001      | 2025-01-02 | 14000       | 340    | 48.20 | 15          |
| CMP003      | 2025-01-01 | 5000        | 60     | 15.00 | 3           |
| CMP002      | 2025-01-02 | 8500        | 150    | 31.00 | 5           |

---

# 🎯 Yêu cầu bài toán

Bạn phải xây dựng một ứng dụng console (CLI) bằng bất kỳ ngôn ngữ nào (Python, NodeJS, Go, Java, Rust, v.v.) để xử lý file CSV và tạo ra dữ liệu analytics đã được aggregate.

---

# 1. Aggregate dữ liệu theo `campaign_id`

Đối với mỗi `campaign_id`, hãy tính:

- `total_impressions`
- `total_clicks`
- `total_spend`
- `total_conversions`
- `CTR` = total_clicks / total_impressions
- `CPA` = total_spend / total_conversions
    - Nếu conversions = 0, hãy bỏ qua hoặc trả về `null` cho CPA

---

# 2. Sinh ra hai danh sách kết quả

## A. Top 10 campaign có CTR cao nhất

Xuất dưới định dạng CSV.

## Format output mong đợi (`top10_ctr.csv`)

| campaign_id | total_impressions | total_clicks | total_spend | total_conversions | CTR    | CPA   |
| ----------- | ----------------- | ------------ | ----------- | ----------------- | ------ | ----- |
| CMP042      | 125000            | 6250         | 12500.50    | 625               | 0.0500 | 20.00 |
| CMP015      | 340000            | 15300        | 30600.25    | 1530              | 0.0450 | 20.00 |
| CMP008      | 890000            | 35600        | 71200.75    | 3560              | 0.0400 | 20.00 |
| CMP023      | 445000            | 15575        | 31150.00    | 1557              | 0.0350 | 20.00 |
| CMP031      | 670000            | 20100        | 40200.50    | 2010              | 0.0300 | 20.00 |

## B. Top 10 campaign có CPA thấp nhất

Xuất dưới định dạng CSV.

Loại bỏ các campaign có conversions bằng 0.

## Format output mong đợi (`top10_cpa.csv`)

| campaign_id | total_impressions | total_clicks | total_spend | total_conversions | CTR    | CPA   |
| ----------- | ----------------- | ------------ | ----------- | ----------------- | ------ | ----- |
| CMP007      | 450000            | 13500        | 13500.00    | 1350              | 0.0300 | 10.00 |
| CMP019      | 780000            | 23400        | 23400.00    | 2340              | 0.0300 | 10.00 |
| CMP033      | 290000            | 8700         | 10440.00    | 870               | 0.0300 | 12.00 |
| CMP012      | 560000            | 16800        | 21840.00    | 1680              | 0.0300 | 13.00 |
| CMP025      | 320000            | 9600         | 13440.00    | 960               | 0.0300 | 14.00 |

---

# 3. Yêu cầu kỹ thuật

- File dữ liệu khá lớn (~1GB).
  Giải pháp của bạn phải xử lý dữ liệu lớn hiệu quả với hiệu năng tốt và tối ưu bộ nhớ.

- Chương trình phải chạy được bằng CLI, ví dụ:

```bash
python aggregator.py --input ad_data.csv --output results/
```

---

# 📬 Hướng dẫn nộp bài

Vui lòng gửi link GitHub repository của bạn qua email đến: [backoffice@flinters.vn](mailto:backoffice@flinters.vn)

Repository của bạn cần bao gồm:

1. Source code trong GitHub repository

2. Các file kết quả:
    - `top10_ctr.csv`
    - `top10_cpa.csv`

3. Một file `README.md` bao gồm:
    - Hướng dẫn setup
    - Cách chạy chương trình
    - Các thư viện sử dụng
    - Thời gian xử lý file 1GB
    - Peak memory usage (nếu có đo)

4. (Không bắt buộc nhưng khuyến khích)
    - Dockerfile
    - Benchmark logs

5. (Nếu có dùng AI) `PROMPTS.md` — xem phần “AI Coding Assistants” bên dưới

---

# ✅ Kỳ vọng về chất lượng code

Vui lòng viết code cẩn thận. Chúng tôi mong đợi:

- Kết quả chính xác — output phải khớp với giá trị mong đợi
- Code sạch, dễ đọc — tên biến/hàm có ý nghĩa, style nhất quán, không có dead code hoặc code bị comment bỏ lại
- Error handling — xử lý tốt các trường hợp thiếu file, malformed rows, edge cases
- Có ý thức về hiệu năng — input ~1GB nên solution phải tối ưu bộ nhớ
- Tests — có test để verify tính đúng đắn của solution
- Có tài liệu giải thích quyết định thiết kế — mô tả ngắn gọn trong README

---

# 🤖 AI Coding Assistants

Chúng tôi khuyến khích bạn sử dụng AI coding assistants như GitHub Copilot, Claude (Cursor AI, Cline), ChatGPT hoặc bất kỳ công cụ AI nào bạn thích.

## Nếu bạn sử dụng AI coding assistants

Hãy thêm file `PROMPTS.md` ở root của repository.

Điều này giúp chúng tôi hiểu:

- cách bạn chia nhỏ vấn đề
- cách bạn giao tiếp với AI tools
- cách bạn tư duy giải quyết vấn đề

## Yêu cầu đối với `PROMPTS.md`

- File phải được đặt tên chính xác là `PROMPTS.md`
- Paste prompt nguyên bản — raw, không chỉnh sửa, đúng như bạn đã nhập
- Không được làm sạch, viết lại hoặc chỉnh sửa prompt trước khi submit

Điều này không bắt buộc nhưng được đánh giá rất cao vì thể hiện khả năng tận dụng các công cụ phát triển hiện đại.

---

Chúc may mắn và coding vui vẻ!
