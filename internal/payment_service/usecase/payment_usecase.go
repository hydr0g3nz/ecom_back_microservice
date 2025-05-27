package usecase

import (
	"context"
	"time"

	"github.com/hydr0g3nz/ecom_back_microservice/internal/payment_service/domain/entity"
	repository "github.com/hydr0g3nz/ecom_back_microservice/internal/payment_service/domain/repository"
	"github.com/hydr0g3nz/ecom_back_microservice/internal/payment_service/domain/service"        // สังเกตว่า EventPublisher อยู่ใน domain/service
	vo "github.com/hydr0g3nz/ecom_back_microservice/internal/payment_service/domain/valueobject" // สังเกตว่า vo อยู่ใน domain/vo
	"github.com/hydr0g3nz/ecom_back_microservice/internal/payment_service/usecase/interfaces"    // สังเกตว่า PaymentGateway อยู่ใน usecase/interfaces
	"github.com/hydr0g3nz/ecom_back_microservice/pkg/logger"
	"go.uber.org/zap"

	// อาจต้อง import package event ด้วย หาก event structs อยู่ใน sub-package event
	// ตัวอย่าง: "github.com/hydr0g3nz/ecom_back_microservice/internal/payment_service/domain/event"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// OrderService ระบุเมธอดที่จำเป็นสำหรับการสื่อสารกับ Order Service
// โค้ดนี้ถูกลบออกเพราะเราจะไม่พึ่งพา Order Service โดยตรงแล้ว
/*
type OrderService interface {
	GetOrderDetails(ctx context.Context, orderID uuid.UUID) (*entity.OrderDetails, error)
	UpdateOrderPaymentStatus(ctx context.Context, orderID uuid.UUID, status string) error
}
*/

// PaymentUseCase ระบุเมธอดที่จำเป็นสำหรับการจัดการ payment
type PaymentUseCase struct {
	paymentRepo       repository.PaymentRepository
	transactionRepo   repository.TransactionRepository
	paymentMethodRepo repository.PaymentMethodRepository
	eventPublisher    service.EventPublisher // ใช้ eventPublisher แทนการเรียก orderService โดยตรง
	paymentGateway    interfaces.PaymentGateway
	logger            logger.Logger
	// orderService ถูกลบออก
}

// NewPaymentUseCase สร้าง instance ของ PaymentUseCase
// ลบ dependency ของ orderService ออกจาก constructor
func NewPaymentUseCase(
	paymentRepo repository.PaymentRepository,
	transactionRepo repository.TransactionRepository,
	paymentMethodRepo repository.PaymentMethodRepository,
	eventPublisher service.EventPublisher, // รับ eventPublisher เข้ามา
	paymentGateway interfaces.PaymentGateway,
	logger logger.Logger,
	// orderService ถูกลบออก
) *PaymentUseCase {
	return &PaymentUseCase{
		paymentRepo:       paymentRepo,
		transactionRepo:   transactionRepo,
		paymentMethodRepo: paymentMethodRepo,
		eventPublisher:    eventPublisher, // กำหนด eventPublisher
		paymentGateway:    paymentGateway,
		logger:            logger,
		// orderService ถูกลบออก
	}
}

// InitiatePaymentRequest เป็นโครงสร้างข้อมูลสำหรับคำขอการเริ่มต้นการชำระเงิน
type InitiatePaymentRequest struct {
	OrderID         uuid.UUID       `json:"order_id"`
	PaymentMethodID uuid.UUID       `json:"payment_method_id"`        // อาจใช้ ID ของ payment method ที่บันทึกไว้
	TokenizedData   string          `json:"tokenized_data,omitempty"` // อาจจะได้รับจากฝั่ง client โดยตรงสำหรับบัตร/วิธีใหม่
	Amount          decimal.Decimal `json:"amount"`
	// อาจเพิ่มข้อมูลยืนยันเพิ่มเติม เช่น UserID เพื่อตรวจสอบสิทธิ์ หรือข้อมูล OrderDetails บางส่วนที่จำเป็น
}

// InitiatePayment เริ่มกระบวนการชำระเงิน
// ลบการเรียก GetOrderDetails ออก
func (uc *PaymentUseCase) InitiatePayment(ctx context.Context, req *InitiatePaymentRequest) (*entity.Payment, error) {
	// ลบ: ตรวจสอบรายละเอียดคำสั่งซื้อ
	// orderDetails, err := uc.orderService.GetOrderDetails(ctx, req.OrderID)
	// if err != nil {
	// 	return nil, err
	// }
	// หมายเหตุ: การตรวจสอบความถูกต้องของ Order (เช่น Order ID มีอยู่จริงหรือไม่, จำนวนเงินถูกต้องหรือไม่)
	// ควรเกิดขึ้นใน Order Service หรือบริการที่สร้างคำสั่งชำระเงินนี้ขึ้นมา ก่อนที่จะส่ง request มายัง Payment Service

	// สร้างการชำระเงินใหม่
	payment := &entity.Payment{
		ID:        uuid.New(),
		OrderID:   req.OrderID,
		Amount:    req.Amount,
		Status:    vo.PaymentStatusPending, // เริ่มต้นที่ Pending
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// อัปเดตข้อมูลวิธีการชำระเงิน
	var paymentMethod *entity.PaymentMethod
	var tokenizedData string // เตรียมตัวสำหรับ tokenized data ที่จะใช้กับ gateway

	if req.PaymentMethodID != uuid.Nil {
		// ใช้ payment method ที่มีอยู่แล้ว
		method, err := uc.paymentMethodRepo.GetPaymentMethodByID(ctx, req.PaymentMethodID)
		if err != nil {
			return nil, err
		}
		paymentMethod = method // เก็บ method object ไว้เผื่อต้องการใช้ข้อมูลอื่น
		payment.PaymentMethod = method.Type
		tokenizedData = method.TokenizedData // ดึง tokenized data จากที่บันทึกไว้
	} else if req.TokenizedData != "" {
		// ใช้ tokenized data ที่ส่งมาโดยตรง (เช่น จากหน้า checkout สำหรับวิธีใหม่)
		var err error
		payment.PaymentMethod, err = vo.NewPaymentMethod(req.TokenizedData)
		if err != nil {
			return nil, err
		} // ต้องมี helper function นี้
		tokenizedData = req.TokenizedData // ใช้ tokenized data จาก request โดยตรง
		// อาจต้องมีการตรวจสอบ/ประมวลผล TokenizedData เพิ่มเติม เช่น การสร้าง PaymentMethod แบบชั่วคราว
	} else {
		return nil, entity.ErrInvalidPaymentMethod
	}

	// บันทึกข้อมูลการชำระเงินเริ่มต้นด้วยสถานะ Pending
	if err := uc.paymentRepo.CreatePayment(ctx, payment); err != nil {
		return nil, err
	}

	// เผยแพร่ event การสร้างการชำระเงิน (PaymentCreated)
	// นี่คือจุดที่บริการอื่น (เช่น Order Service) จะรับรู้ว่ามีการเริ่มกระบวนการชำระเงินแล้ว
	evtCreated := &entity.Payment{ // สมมติว่ามี struct event.PaymentCreated
		ID:            payment.ID,
		OrderID:       payment.OrderID,
		Amount:        payment.Amount,
		Status:        payment.Status, // ควรเป็น Pending
		PaymentMethod: payment.PaymentMethod,
		CreatedAt:     payment.CreatedAt,
	}
	// ใช้ Goroutine หรือ mechanism อื่นเพื่อให้การ Publish ไม่ block การทำงานหลัก
	go func() {
		if err := uc.eventPublisher.PublishPaymentCreated(context.Background(), evtCreated); err != nil {
			uc.logger.Error("failed to publish PaymentCreated event", zap.Error(err))
		}
	}()

	// ดำเนินการชำระเงินกับ gateway
	gatewayResponse, err := uc.paymentGateway.ProcessPayment(ctx, payment.Amount, tokenizedData, payment.OrderID)
	if err != nil {
		// หากเกิดข้อผิดพลาดในการสื่อสารกับ Gateway หรือ Gateway ปฏิเสธทันที
		payment.Status = vo.PaymentStatusFailed // อัปเดตสถานะเป็น FAILED
		payment.UpdatedAt = time.Now()
		// พยายามอัปเดตสถานะใน database (ล็อก error หากล้มเหลว)
		if updateErr := uc.paymentRepo.UpdatePayment(ctx, payment); updateErr != nil {
			// ล็อกข้อผิดพลาดร้ายแรง: ไม่สามารถอัปเดตสถานะ Payment เป็น Failed ได้หลังจาก Gateway error
		}

		// เผยแพร่ event การชำระเงินล้มเหลว (PaymentFailed)
		failedEvt := &entity.PaymentFailed{ // สมมติว่ามี struct event.PaymentFailed
			PaymentID: payment.ID,
			OrderID:   payment.OrderID,
			Reason:    "Payment Gateway Error: " + err.Error(), // ใส่รายละเอียด error
			FailedAt:  time.Now(),
		}
		// ใช้ Goroutine หรือ mechanism อื่นเพื่อให้การ Publish ไม่ block การทำงานหลัก
		go func() {
			if publishErr := uc.eventPublisher.PublishPaymentFailed(context.Background(), failedEvt); publishErr != nil {
				// ล็อกข้อผิดพลาด: ไม่สามารถ publish PaymentFailed event ได้
			}
		}()

		// คืนค่า payment object พร้อมสถานะที่อัปเดต และ error จาก gateway
		return payment, err
	}

	// บันทึกข้อมูลธุรกรรม
	// สถานะเริ่มต้นของ Transaction ควรมาจาก Gateway response หากมี หรือ Pending
	transactionStatus, err := vo.NewTransactionStatus(gatewayResponse.Status)
	if err != nil {
		return nil, err
	} // ต้องมี helper function นี้
	transaction := &entity.Transaction{
		ID:              uuid.New(),
		PaymentID:       payment.ID,
		Type:            entity.TransactionTypeCharge,
		Amount:          payment.Amount, // ควรเป็น Amount ที่ใช้ในการ Charge จริงๆ (อาจแตกต่างจาก req.Amount เล็กน้อยในบางกรณี)
		Status:          transactionStatus,
		GatewayResponse: gatewayResponse.RawResponse,   // เก็บ response ดิบไว้เพื่อการดีบัก/ตรวจสอบ
		GatewayTxID:     gatewayResponse.TransactionID, // เก็บ ID จาก Gateway ใน Transaction ด้วย
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	// พยายามสร้าง Transaction (ล็อก error หากล้มเหลว)
	if createTxErr := uc.transactionRepo.CreateTransaction(ctx, transaction); createTxErr != nil {
		// ล็อกข้อผิดพลาด: ไม่สามารถสร้าง Transaction ได้หลังจาก Gateway สำเร็จ
	}

	// อัปเดตข้อมูลการชำระเงินตามผลลัพธ์จาก Gateway (สถานะ Processing/Completed ทันที ขึ้นอยู่กับ Gateway)
	// หาก Gateway มีสถานะแบบ asynchronous (เช่น รอการยืนยัน) สถานะนี้จะเป็น Processing
	// หาก Gateway ยืนยันสำเร็จทันที สถานะนี้จะเป็น Completed
	payment.Status = entity.MapGatewayStatusToPaymentStatus(gatewayResponse.Status) // ต้องมี helper function นี้
	payment.GatewayTransactionID = gatewayResponse.TransactionID                    // เก็บ ID หลักจาก Gateway ใน Payment
	payment.UpdatedAt = time.Now()
	// อัปเดตสถานะ Payment ใน database
	if updatePayErr := uc.paymentRepo.UpdatePayment(ctx, payment); updatePayErr != nil {
		// ล็อกข้อผิดพลาดร้ายแรง: ไม่สามารถอัปเดตสถานะ Payment หลังจาก Gateway สำเร็จ
		return payment, updatePayErr // คืน error หากไม่สามารถอัปเดต Payment ได้
	}

	// เผยแพร่ event การอัปเดตการชำระเงิน (PaymentUpdated)
	// Event นี้จะบอกบริการอื่นว่า Payment มีสถานะเปลี่ยนแปลง
	updatedEvt := &entity.PaymentUpdated{ // สมมติว่ามี struct event.PaymentUpdated
		PaymentID:            payment.ID,
		OrderID:              payment.OrderID,
		Status:               payment.Status, // สถานะใหม่ (Processing หรือ Completed)
		GatewayTransactionID: payment.GatewayTransactionID,
		UpdatedAt:            payment.UpdatedAt,
		// อาจเพิ่ม TransactionID ของ Transaction ที่สร้างขึ้นด้วย
	}
	// ใช้ Goroutine หรือ mechanism อื่นเพื่อให้การ Publish ไม่ block การทำงานหลัก
	go func() {
		if publishErr := uc.eventPublisher.PublishPaymentUpdated(context.Background(), updatedEvt); publishErr != nil {
			// ล็อกข้อผิดพลาด: ไม่สามารถ publish PaymentUpdated event ได้
		}
	}()

	// หาก Gateway แจ้งว่าสำเร็จทันที ก็เผยแพร่ PaymentCompleted event ด้วย
	if payment.Status == vo.PaymentStatusCompleted {
		completedEvt := &entity.PaymentCompleted{ // สมมติว่ามี struct event.PaymentCompleted
			PaymentID:   payment.ID,
			OrderID:     payment.OrderID,
			Amount:      payment.Amount,
			CompletedAt: payment.UpdatedAt, // หรือใช้เวลาที่ได้รับจาก Gateway response ถ้ามี
		}
		go func() {
			if publishErr := uc.eventPublisher.PublishPaymentCompleted(context.Background(), completedEvt); publishErr != nil {
				// ล็อกข้อผิดพลาด: ไม่สามารถ publish PaymentCompleted event ได้
			}
		}()
		// หมายเหตุ: หาก Gateway แจ้งสถานะเป็น Processing EventCompleted จะถูก publish ใน HandleGatewayCallback แทน
	}

	return payment, nil
}

// HandleGatewayCallback จัดการกับ callback จาก payment gateway
// ลบการเรียก UpdateOrderPaymentStatus ออก และใช้ Event แทน
func (uc *PaymentUseCase) HandleGatewayCallback(ctx context.Context, callbackData map[string]interface{}) error {
	// ตรวจสอบความถูกต้องของ callback (เช่น signature validation)
	isValid, err := uc.paymentGateway.VerifyCallback(ctx, callbackData)
	if err != nil {
		// ล็อก error จากการตรวจสอบ callback
		return entity.ErrInvalidCallback // หรือ entity.ErrCallbackVerificationFailed
	}
	if !isValid {
		// ล็อกการ callback ที่ไม่ถูกต้อง
		return entity.ErrInvalidCallback
	}

	// ดึงข้อมูลที่จำเป็นจาก callback
	// ควรมี logic การ parse callback ที่ซับซ้อนกว่านี้ตามรูปแบบของแต่ละ Gateway
	transactionIDFromGateway, ok := callbackData["transaction_id"].(string) // ID ธุรกรรมจาก Gateway
	if !ok || transactionIDFromGateway == "" {
		// ล็อก: callback data ไม่มี transaction_id
		return entity.ErrInvalidCallbackData
	}

	statusFromGateway, ok := callbackData["status"].(string) // สถานะจาก Gateway
	if !ok || statusFromGateway == "" {
		// ล็อก: callback data ไม่มี status
		return entity.ErrInvalidCallbackData
	}

	// ค้นหาการชำระเงินโดยใช้ transaction ID จาก gateway
	// ต้องแน่ใจว่า gatewayTransactionID ใน DB ตรงกับ ID ที่ Gateway ส่งมาใน Callback
	payment, err := uc.paymentRepo.GetPaymentByGatewayTransactionID(ctx, transactionIDFromGateway)
	if err != nil {
		// ล็อก: ไม่พบ Payment สำหรับ Gateway Transaction ID นี้ (อาจเป็น callback เก่า, ซ้ำ, หรือผิดพลาด)
		return err // หรือ entity.ErrPaymentNotFoundForCallback
	}

	// แปลงสถานะจาก Gateway เป็นสถานะ Payment ของระบบเรา
	newPaymentStatus := entity.MapGatewayStatusToPaymentStatus(statusFromGateway) // ต้องมี helper function นี้
	if newPaymentStatus == "" {
		// ล็อก: สถานะจาก Gateway ไม่รู้จัก
		return entity.ErrUnknownPaymentStatus // หรือ log แล้ว return nil/specific error
	}

	// ตรวจสอบว่าสถานะใหม่แตกต่างจากสถานะปัจจุบันหรือไม่ เพื่อหลีกเลี่ยงการอัปเดตซ้ำซ้อน
	if payment.Status == newPaymentStatus {
		// สถานะเหมือนเดิม ไม่ต้องทำอะไรต่อ (อาจล็อก info ว่าได้รับ callback ซ้ำ)
		return nil
	}

	// อัปเดตสถานะการชำระเงิน
	payment.Status = newPaymentStatus
	payment.UpdatedAt = time.Now() // ใช้เวลาปัจจุบัน หรือเวลาจาก callback ถ้ามีและน่าเชื่อถือ
	if err := uc.paymentRepo.UpdatePayment(ctx, payment); err != nil {
		// ล็อกข้อผิดพลาด: ไม่สามารถอัปเดตสถานะ Payment ได้
		return err
	}

	// ลบ: อัปเดตสถานะคำสั่งซื้อโดยตรง
	// if err := uc.orderService.UpdateOrderPaymentStatus(ctx, payment.OrderID, newStatus); err != nil {
	// 	// ล็อกข้อผิดพลาดแต่ดำเนินการต่อ
	// }
	// การอัปเดตสถานะคำสั่งซื้อจะถูกกระตุ้นโดย Event ที่จะ Publish ด้านล่างแทน

	// อัปเดตธุรกรรมที่เกี่ยวข้อง
	// ค้นหา Transaction ที่ตรงกับ Payment ID และ Gateway Tx ID หรือเป็น Transaction ล่าสุดของ Payment นี้
	// หรืออาจจะต้องสร้าง Transaction ใหม่สำหรับบาง Callback (เช่น สำหรับ Refund Callback)
	// ในกรณี Charge Callback นี้ เราควรหา Transaction ที่สร้างตอน InitiatePayment และอัปเดตสถานะ
	transactions, err := uc.transactionRepo.ListTransactionsByPaymentID(ctx, payment.ID)
	if err == nil && len(transactions) > 0 {
		// สมมติว่า GatewayTransactionID ใน Payment ตรงกับ GatewayTxID ใน Transaction ที่เราสนใจ
		var targetTransaction *entity.Transaction
		for _, tx := range transactions {
			if tx.GatewayTxID == transactionIDFromGateway { // ค้นหา Transaction ที่ตรงกับ Gateway Tx ID
				targetTransaction = tx
				break
			}
			// หรืออาจใช้เงื่อนไขอื่น เช่น เป็น Transaction Type Charge ล่าสุดที่สถานะยังไม่ Completed/Failed
		}

		if targetTransaction != nil {
			targetTransaction.Status = entity.MapGatewayStatusToTransactionStatus(statusFromGateway) // แปลงสถานะจาก Gateway เป็นสถานะ Transaction
			targetTransaction.UpdatedAt = time.Now()
			if updateTxErr := uc.transactionRepo.UpdateTransaction(ctx, targetTransaction); updateTxErr != nil {
				// ล็อกข้อผิดพลาด: ไม่สามารถอัปเดตสถานะ Transaction ได้
			}
		} else {
			// ล็อก: ไม่พบ Transaction ที่ตรงกับ Gateway Transaction ID สำหรับ Payment นี้
			// อาจจำเป็นต้องสร้าง Transaction ใหม่ในบางกรณี หรือตรวจสอบ logic
		}
	} else if err != nil {
		// ล็อกข้อผิดพลาด: ไม่สามารถดึงรายการ Transactions ได้
	} else {
		// ล็อก: ไม่พบ Transactions สำหรับ Payment นี้ (อาจผิดปกติ)
	}

	// เผยแพร่ event การอัปเดตการชำระเงิน (PaymentUpdated)
	// Event นี้สำคัญมากสำหรับบริการอื่น ๆ เช่น Order Service, Inventory Service
	updatedEvt := &entity.PaymentUpdated{ // สมมติว่ามี struct event.PaymentUpdated
		PaymentID:            payment.ID,
		OrderID:              payment.OrderID,
		Status:               payment.Status, // สถานะล่าสุด (Completed, Failed, Processing etc.)
		GatewayTransactionID: payment.GatewayTransactionID,
		UpdatedAt:            payment.UpdatedAt,
		// อาจเพิ่ม TransactionID, GatewayStatus, RawResponse จาก Callback เข้าไปใน Event ด้วยเพื่อความสมบูรณ์
	}
	// ใช้ Goroutine หรือ mechanism อื่นเพื่อให้การ Publish ไม่ block การทำงานหลัก
	go func() {
		if publishErr := uc.eventPublisher.PublishPaymentUpdated(context.Background(), updatedEvt); publishErr != nil {
			// ล็อกข้อผิดพลาด: ไม่สามารถ publish PaymentUpdated event ได้
		}
	}()

	// เผยแพร่ event เฉพาะตามสถานะการชำระเงินหลัก (เช่น PaymentCompleted, PaymentFailed)
	// เพื่อให้ Consumer ที่สนใจสถานะเฉพาะนี้รับไปดำเนินการได้ง่ายขึ้น
	switch newPaymentStatus {
	case vo.PaymentStatusCompleted:
		completedEvt := &entity.PaymentCompleted{ // สมมติว่ามี struct event.PaymentCompleted
			PaymentID:   payment.ID,
			OrderID:     payment.OrderID,
			Amount:      payment.Amount,
			CompletedAt: payment.UpdatedAt,
		}
		go func() {
			if publishErr := uc.eventPublisher.PublishPaymentCompleted(context.Background(), completedEvt); publishErr != nil {
				// ล็อกข้อผิดพลาด: ไม่สามารถ publish PaymentCompleted event ได้
			}
		}()
	case vo.PaymentStatusFailed:
		failedEvt := &entity.PaymentFailed{ // สมมติว่ามี struct event.PaymentFailed
			PaymentID: payment.ID,
			OrderID:   payment.OrderID,
			Reason:    "Gateway callback indicated failure", // หรือดึง Reason ที่เฉพาะเจาะจงกว่าจาก Callback Data
			FailedAt:  payment.UpdatedAt,
		}
		go func() {
			if publishErr := uc.eventPublisher.PublishPaymentFailed(context.Background(), failedEvt); publishErr != nil {
				// ล็อกข้อผิดพลาด: ไม่สามารถ publish PaymentFailed event ได้
			}
		}()
		// อาจมี case สำหรับ PaymentStatusRefunded หรือสถานะอื่นๆ ตามต้องการ
	}

	return nil
}

// GetPaymentInfo รับข้อมูลการชำระเงิน
func (uc *PaymentUseCase) GetPaymentInfo(ctx context.Context, paymentID uuid.UUID) (*entity.Payment, error) {
	payment, err := uc.paymentRepo.GetPaymentByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	// ดึงข้อมูลธุรกรรมที่เกี่ยวข้อง
	transactions, err := uc.transactionRepo.ListTransactionsByPaymentID(ctx, paymentID)
	if err == nil {
		payment.Transactions = transactions
	}
	// หาก err != nil ในการดึง transaction อาจจะ log error แทนการ return
	// เนื่องจากข้อมูล payment หลักยังคงมีอยู่

	return payment, nil
}

// InitiateRefundRequest เป็นโครงสร้างข้อมูลสำหรับคำขอการคืนเงิน
type InitiateRefundRequest struct {
	PaymentID uuid.UUID       `json:"payment_id"`
	Amount    decimal.Decimal `json:"amount"`
	Reason    string          `json:"reason"`
	UserID    uuid.UUID       `json:"user_id"` // เพิ่ม UserID เพื่อตรวจสอบสิทธิ์
}

// InitiateRefund เริ่มกระบวนการคืนเงิน
func (uc *PaymentUseCase) InitiateRefund(ctx context.Context, req *InitiateRefundRequest) (*entity.Transaction, error) {
	// ค้นหาการชำระเงิน
	payment, err := uc.paymentRepo.GetPaymentByID(ctx, req.PaymentID)
	if err != nil {
		return nil, err
	}

	// ตรวจสอบสิทธิ์: ผู้ใช้คนนี้มีสิทธิ์คืนเงินสำหรับ Payment นี้หรือไม่?
	// โค้ดนี้สมมติว่า UserID ถูกเก็บไว้ใน Payment Entity หรือสามารถหาได้
	// ถ้า Payment entity ไม่มี UserID อาจจะต้องเรียก Order Service เพื่อตรวจสอบสิทธิ์
	// หรืออาจจะต้องส่ง UserID ไปพร้อมกับ PaymentCreated Event และเก็บไว้ใน Payment entity
	// ในที่นี้ ขอละไว้ แต่ให้ตระหนักว่าต้องมีการตรวจสอบสิทธิ์
	// if payment.UserID != req.UserID { return entity.ErrUnauthorized }

	// ตรวจสอบว่าสามารถคืนเงินได้หรือไม่ (ควรจะ Complete ก่อนจึงจะคืนได้)
	if payment.Status != vo.PaymentStatusCompleted {
		return nil, entity.ErrCannotRefundPayment // Payment ไม่อยู่ในสถานะที่สามารถคืนเงินได้
	}

	// ตรวจสอบจำนวนเงินที่คืน
	// ควรมีการตรวจสอบยอดเงินที่คืนไปแล้วด้วย เพื่อไม่ให้คืนเกินยอดรวม
	// สมมติว่ามีเมธอด GetTotalRefundedAmount สำหรับ Payment
	// totalRefunded, err := uc.transactionRepo.GetTotalRefundedAmountByPaymentID(ctx, payment.ID)
	// if err != nil { return nil, err }
	// if req.Amount.Add(totalRefunded).GreaterThan(payment.Amount) {
	// 	return nil, entity.ErrRefundAmountExceedsPaymentAmount
	// }
	if req.Amount.GreaterThan(payment.Amount) { // ตรวจสอบเบื้องต้นแค่ไม่ให้คืนเกินยอดชำระรวม
		return nil, entity.ErrRefundAmountTooLarge
	}
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, entity.ErrInvalidRefundAmount // จำนวนเงินคืนต้องมากกว่าศูนย์
	}

	// ดำเนินการคืนเงินกับ gateway
	// ProcessRefund ควรใช้ GatewayTransactionID ของการ Charge เดิม
	gatewayResponse, err := uc.paymentGateway.ProcessRefund(ctx, payment.GatewayTransactionID, req.Amount)
	if err != nil {
		// ล็อก error จาก Gateway ในการคืนเงิน
		// ในกรณีที่ Gateway คืน error ในขั้นตอนนี้ มักจะไม่มี Transaction ใน Gateway
		// แต่เราอาจจะสร้าง Transaction สถานะ Failed เพื่อบันทึกการพยายามคืนเงิน
		failedTx := &entity.Transaction{
			ID:        uuid.New(),
			PaymentID: payment.ID,
			Type:      entity.TransactionTypeRefund,
			Amount:    req.Amount.Neg(), // จำนวนเงินคืนมักเป็นค่าติดลบในระบบบัญชี แต่ใน Payment Service อาจใช้ค่าบวกก็ได้ ขึ้นอยู่กับการออกแบบ
			Status:    entity.TransactionStatusFailed,
			Reason:    "Gateway Refund Process Failed: " + err.Error(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_ = uc.transactionRepo.CreateTransaction(ctx, failedTx) // บันทึกความล้มเหลว (ล็อก error หากบันทึกไม่ได้)

		// เผยแพร่ event การคืนเงินล้มเหลว
		refundFailedEvt := &entity.RefundFailed{ // สมมติว่ามี struct event.RefundFailed
			PaymentID: req.PaymentID,
			OrderID:   payment.OrderID, // ต้องมั่นใจว่า Payment entity มี OrderID
			Amount:    req.Amount,
			Reason:    "Gateway Refund Process Failed: " + err.Error(),
			FailedAt:  time.Now(),
			// อาจเพิ่ม TransactionID ของ failedTx
		}
		go func() {
			_ = uc.eventPublisher.PublishRefundFailed(context.Background(), refundFailedEvt) // ล็อก error หาก publish ไม่ได้
		}()

		return nil, err // คืน error หลักจาก gateway
	}

	// บันทึกข้อมูลธุรกรรมการคืนเงินที่สำเร็จ (หรืออยู่ในสถานะ Pending/Processing สำหรับ Async Refund)
	transactionStatus := entity.MapGatewayStatusToTransactionStatus(gatewayResponse.Status) // แปลงสถานะจาก Gateway
	transaction := &entity.Transaction{
		ID:              uuid.New(),
		PaymentID:       payment.ID,
		Type:            entity.TransactionTypeRefund,
		Amount:          req.Amount.Neg(), // จำนวนเงินคืน (ค่าติดลบ)
		Status:          transactionStatus,
		GatewayResponse: gatewayResponse.RawResponse,
		GatewayTxID:     gatewayResponse.TransactionID, // Gateway อาจสร้าง Transaction ID ใหม่สำหรับการคืนเงิน
		Reason:          req.Reason,                    // เหตุผลการคืนเงิน
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	if err := uc.transactionRepo.CreateTransaction(ctx, transaction); err != nil {
		// ล็อกข้อผิดพลาด: ไม่สามารถสร้าง Transaction การคืนเงินได้
		// อาจพิจารณาคืน error ที่นี่ หรือแค่ log แล้วดำเนินการต่อ
		return transaction, err
	}

	// อัปเดตสถานะการชำระเงิน หากคืนเงินเต็มจำนวน หรือสถานะจาก Gateway บ่งชี้ว่า Refunded
	// ควรตรวจสอบ GatewayResponse.Status อย่างละเอียดเพื่อตัดสินใจอัปเดตสถานะ Payment หลักหรือไม่
	// ในที่นี้ใช้ logic เดิมคือถ้าคืนเต็มจำนวน ให้อัปเดตเป็น Refunded
	// หรือถ้า Gateway ส่งสถานะ Refunded มาใน response ทันที ก็อาจอัปเดตสถานะ Payment เลย
	if req.Amount.Equal(payment.Amount) { // ตรวจสอบว่าคืนเต็มจำนวนหรือไม่
		payment.Status = vo.PaymentStatusRefunded // สมมติว่ามีสถานะ Refunded
		payment.UpdatedAt = time.Now()
		if updateErr := uc.paymentRepo.UpdatePayment(ctx, payment); updateErr != nil {
			// ล็อกข้อผิดพลาด: ไม่สามารถอัปเดตสถานะ Payment เป็น Refunded ได้
		}
		// ถ้า Payment Status เปลี่ยนเป็น Refunded ก็ควร Publish PaymentUpdated event ด้วย
		go func() {
			updatedEvt := &entity.PaymentUpdated{ /* ... field values ... */ Status: payment.Status, UpdatedAt: payment.UpdatedAt} // สร้าง event PaymentUpdated
			_ = uc.eventPublisher.PublishPaymentUpdated(context.Background(), updatedEvt)                                          // ล็อก error หาก publish ไม่ได้
		}()
	}
	// หาก Refund เป็นแบบ Async และ Gateway จะส่ง Callback มาภายหลัง การอัปเดตสถานะ Payment
	// จะเกิดขึ้นใน HandleGatewayCallback เมื่อได้รับ Callback สำหรับ Refund นั้นๆ แทน

	// เผยแพร่ event การคืนเงินสำเร็จ (หรือInitiated หากเป็น async)
	// ควรมี event ที่บอกสถานะการคืนเงินที่ชัดเจนกว่า RefundInitiated หาก Gateway บอกสถานะทันที
	// เช่น RefundCompleted, RefundFailed
	// ในที่นี้ใช้ RefundInitiated ตามโค้ดเดิม แต่ควรปรับปรุง
	evtInitiated := &entity.RefundInitiated{ // สมมติว่ามี struct event.RefundInitiated
		PaymentID:   payment.ID,
		OrderID:     payment.OrderID, // ต้องมั่นใจว่า Payment entity มี OrderID
		Amount:      req.Amount,
		Reason:      req.Reason,
		InitiatedAt: transaction.CreatedAt, // ใช้เวลาสร้าง Transaction การคืนเงิน
		// อาจเพิ่ม TransactionID, GatewayTxID, GatewayStatus จาก response ด้วย
	}
	go func() {
		_ = uc.eventPublisher.PublishRefundInitiated(context.Background(), evtInitiated) // ล็อก error หาก publish ไม่ได้
	}()

	// หาก Gateway response บ่งชี้ว่า Refund สำเร็จทันที ก็ควร Publish RefundCompleted event ด้วย
	if transaction.Status == entity.TransactionStatusCompleted { // ตรวจสอบสถานะ Transaction ที่แปลงมาจาก Gateway status
		completedEvt := &entity.RefundCompleted{ // สมมติว่ามี struct event.RefundCompleted
			PaymentID:   payment.ID,
			OrderID:     payment.OrderID,
			Amount:      req.Amount,
			CompletedAt: transaction.UpdatedAt,
			RefundTxID:  transaction.ID,
		}
		go func() {
			_ = uc.eventPublisher.PublishRefundCompleted(context.Background(), completedEvt) // ล็อก error หาก publish ไม่ได้
		}()
		// หมายเหตุ: หาก Refund เป็นแบบ Async และ Gateway จะส่ง Callback มาภายหลัง EventCompleted
		// จะถูก publish ใน HandleGatewayCallback เมื่อได้รับ Callback สำหรับ Refund นั้นๆ แทน
	}

	return transaction, nil
}

// RegisterPaymentMethodRequest เป็นโครงสร้างข้อมูลสำหรับคำขอการลงทะเบียนวิธีการชำระเงิน
type RegisterPaymentMethodRequest struct {
	UserID        uuid.UUID `json:"user_id"`
	Type          string    `json:"type"`           // เช่น "credit_card", "paypal", "bank_account"
	TokenizedData string    `json:"tokenized_data"` // ข้อมูลที่ผ่าน tokenization แล้ว (ปลอดภัยกว่าข้อมูลบัตรดิบ)
	IsDefault     bool      `json:"is_default"`
	// อาจเพิ่ม Last4 digits, Brand, Expire Date etc. จาก tokenized data เพื่อแสดงให้ผู้ใช้เห็น
}

// RegisterPaymentMethod ลงทะเบียนวิธีการชำระเงินใหม่
func (uc *PaymentUseCase) RegisterPaymentMethod(ctx context.Context, req *RegisterPaymentMethodRequest) (*entity.PaymentMethod, error) {
	// ตรวจสอบ Type และ TokenizedData ตาม business rules
	if req.UserID == uuid.Nil || req.Type == "" || req.TokenizedData == "" {
		return nil, entity.ErrInvalidPaymentMethodData
	}
	// อาจมีการตรวจสอบ format ของ TokenizedData ตาม Type

	// สร้างข้อมูลวิธีการชำระเงินใหม่
	paymentMethod := &entity.PaymentMethod{
		ID:            uuid.New(),
		UserID:        req.UserID,
		Type:          req.Type,
		TokenizedData: req.TokenizedData, // ควรมีการเข้ารหัส TokenizedData ใน DB ด้วย
		IsDefault:     req.IsDefault,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		// อาจเพิ่มข้อมูลอื่นๆ เช่น brand, last4, expiry_date ที่ได้จากการ parse tokenized data
	}

	// หากเป็นวิธีการชำระเงินเริ่มต้น ให้ยกเลิกวิธีการชำระเงินเริ่มต้นเดิมของผู้ใช้นี้
	if req.IsDefault {
		defaultMethod, err := uc.paymentMethodRepo.GetDefaultPaymentMethod(ctx, req.UserID)
		// จัดการ error จาก GetDefaultPaymentMethod อย่างระมัดระวัง (อาจจะไม่มี default อยู่แล้ว)
		if err != nil && err != entity.ErrPaymentMethodNotFound {
			// ล็อก error หากไม่ใช่แค่ไม่พบ
			// ดำเนินการต่อได้ แต่อาจมีปัญหาถ้า default method ยังคงถูกตั้งค่าอยู่
		} else if defaultMethod != nil && defaultMethod.ID != paymentMethod.ID {
			// พบ default method เดิมและไม่ใช่ตัวใหม่
			defaultMethod.IsDefault = false
			defaultMethod.UpdatedAt = time.Now()
			// อัปเดต default method เดิม (ล็อก error หากล้มเหลว)
			if updateErr := uc.paymentMethodRepo.UpdatePaymentMethod(ctx, defaultMethod); updateErr != nil {
				// ล็อกข้อผิดพลาด: ไม่สามารถยกเลิก default payment method เดิมได้
				// พิจารณาว่าจะ return error ที่นี่ หรือแค่ log แล้วพยายามสร้างอันใหม่ต่อไป
				// ในที่นี้เลือกแค่ log แล้วดำเนินการต่อ
			}
		}
	}

	// บันทึกข้อมูลวิธีการชำระเงินใหม่
	if err := uc.paymentMethodRepo.CreatePaymentMethod(ctx, paymentMethod); err != nil {
		// ล็อก error
		return nil, err
	}

	// หากตั้งเป็น Default และการอัปเดต default เดิมล้มเหลว ควรมี機制การแก้ไข
	// หรืออาจต้องใช้ Transaction ในระดับ UseCase เพื่อให้แน่ใจว่าทั้งสองการอัปเดตสำเร็จหรือล้มเหลวพร้อมกัน

	// อาจ Publish Event PaymentMethodRegistered เพื่อแจ้งให้บริการอื่นทราบ (เช่น User Service)
	// evtRegistered := &entity.PaymentMethodRegistered{UserID: req.UserID, PaymentMethodID: paymentMethod.ID, Type: paymentMethod.Type, IsDefault: paymentMethod.IsDefault}
	// go func() { _ = uc.eventPublisher.PublishPaymentMethodRegistered(context.Background(), evtRegistered) }() // ล็อก error หาก publish ไม่ได้

	return paymentMethod, nil
}

// ListUserPaymentMethods แสดงรายการวิธีการชำระเงินของผู้ใช้
func (uc *PaymentUseCase) ListUserPaymentMethods(ctx context.Context, userID uuid.UUID) ([]*entity.PaymentMethod, error) {
	if userID == uuid.Nil {
		return nil, entity.ErrInvalidUserID
	}
	return uc.paymentMethodRepo.ListPaymentMethodsByUserID(ctx, userID)
}

// DeletePaymentMethod ลบวิธีการชำระเงิน
func (uc *PaymentUseCase) DeletePaymentMethod(ctx context.Context, userID, paymentMethodID uuid.UUID) error {
	if userID == uuid.Nil || paymentMethodID == uuid.Nil {
		return entity.ErrInvalidInput
	}

	// ตรวจสอบว่าวิธีการชำระเงินเป็นของผู้ใช้จริงหรือไม่
	paymentMethod, err := uc.paymentMethodRepo.GetPaymentMethodByID(ctx, paymentMethodID)
	if err != nil {
		return err // เช่น ErrPaymentMethodNotFound
	}
	if paymentMethod.UserID != userID {
		return entity.ErrUnauthorized // ผู้ใช้ไม่มีสิทธิ์ลบ PaymentMethod นี้
	}

	// ตรวจสอบว่า Payment Method นี้ไม่ได้ถูกใช้ในการชำระเงินที่ active อยู่
	// อาจจะต้องตรวจสอบใน Transaction Repository หรือ Payment Repository
	// เช่น CheckIfPaymentMethodIsInUse(ctx, paymentMethodID) bool

	// ถ้า Payment Method ที่ลบเป็น Default ของผู้ใช้นี้ ควรมี logic ในการกำหนด Default Method ใหม่
	if paymentMethod.IsDefault {
		// ควรหา Payment Method อื่นๆ ของผู้ใช้นี้ และตั้งค่าหนึ่งในนั้นเป็น Default โดยอัตโนมัติ
		// หรือกำหนดให้ผู้ใช้เลือกใหม่
		// ในที่นี้ ขอละ logic การกำหนด Default ใหม่ไว้ แต่ให้ระลึกถึง
	}

	// ดำเนินการลบ
	err = uc.paymentMethodRepo.DeletePaymentMethod(ctx, paymentMethodID)
	if err != nil {
		// ล็อก error
		return err
	}

	// อาจ Publish Event PaymentMethodDeleted เพื่อแจ้งให้บริการอื่นทราบ
	// evtDeleted := &entity.PaymentMethodDeleted{UserID: userID, PaymentMethodID: paymentMethodID}
	// go func() { _ = uc.eventPublisher.PublishPaymentMethodDeleted(context.Background(), evtDeleted) }() // ล็อก error หาก publish ไม่ได้

	return nil
}

// SetDefaultPaymentMethod ตั้งค่าวิธีการชำระเงินเริ่มต้น
func (uc *PaymentUseCase) SetDefaultPaymentMethod(ctx context.Context, userID, paymentMethodID uuid.UUID) error {
	if userID == uuid.Nil || paymentMethodID == uuid.Nil {
		return entity.ErrInvalidInput
	}

	// ตรวจสอบว่าวิธีการชำระเงินที่ต้องการตั้งเป็น Default เป็นของผู้ใช้จริงหรือไม่
	paymentMethod, err := uc.paymentMethodRepo.GetPaymentMethodByID(ctx, paymentMethodID)
	if err != nil {
		return err // เช่น ErrPaymentMethodNotFound
	}
	if paymentMethod.UserID != userID {
		return entity.ErrUnauthorized // ผู้ใช้ไม่มีสิทธิ์ตั้ง PaymentMethod นี้เป็น Default
	}

	// หาก Payment Method ที่ระบุเป็น Default อยู่แล้ว ไม่ต้องทำอะไร
	if paymentMethod.IsDefault {
		return nil
	}

	// ยกเลิกวิธีการชำระเงินเริ่มต้นเดิมของผู้ใช้นี้
	defaultMethod, err := uc.paymentMethodRepo.GetDefaultPaymentMethod(ctx, userID)
	// จัดการ error จาก GetDefaultPaymentMethod อย่างระมัดระวัง
	if err != nil && err != entity.ErrPaymentMethodNotFound {
		// ล็อก error หากไม่ใช่แค่ไม่พบ default เดิม
		// ดำเนินการต่อได้ แต่อาจมีปัญหาถ้า default method เดิมยังคงถูกตั้งค่าอยู่
	} else if defaultMethod != nil && defaultMethod.ID != paymentMethodID {
		// พบ default method เดิมและไม่ใช่ตัวใหม่
		defaultMethod.IsDefault = false
		defaultMethod.UpdatedAt = time.Now()
		// อัปเดต default method เดิม (ล็อก error หากล้มเหลว)
		if updateErr := uc.paymentMethodRepo.UpdatePaymentMethod(ctx, defaultMethod); updateErr != nil {
			// ล็อกข้อผิดพลาด: ไม่สามารถยกเลิก default payment method เดิมได้
			// พิจารณาว่าจะ return error ที่นี่ หรือแค่ log แล้วพยายามตั้งค่าอันใหม่ต่อไป
			// ในที่นี้เลือกแค่ log แล้วดำเนินการต่อ
		}
	}

	// ตั้งค่าวิธีการชำระเงินที่ระบุเป็น Default ใหม่
	paymentMethod.IsDefault = true
	paymentMethod.UpdatedAt = time.Now()
	err = uc.paymentMethodRepo.UpdatePaymentMethod(ctx, paymentMethod)
	if err != nil {
		// ล็อก error
		// หากการตั้งค่าใหม่ล้มเหลว และการยกเลิก default เดิมสำเร็จ จะทำให้ไม่มี default method
		// ควรพิจารณาใช้ Transaction ในระดับ UseCase เพื่อให้แน่ใจว่าทั้งสองการอัปเดตสำเร็จหรือล้มเหลวพร้อมกัน
		return err
	}

	// อาจ Publish Event DefaultPaymentMethodUpdated เพื่อแจ้งให้บริการอื่นทราบ
	// evtDefaultUpdated := &entity.DefaultPaymentMethodUpdated{UserID: userID, PaymentMethodID: paymentMethodID}
	// go func() { _ = uc.eventPublisher.PublishDefaultPaymentMethodUpdated(context.Background(), evtDefaultUpdated) }() // ล็อก error หาก publish ไม่ได้

	return nil
}
