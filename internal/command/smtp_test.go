package command

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/zitadel/zitadel/internal/api/authz"
	"github.com/zitadel/zitadel/internal/crypto"
	"github.com/zitadel/zitadel/internal/domain"
	caos_errs "github.com/zitadel/zitadel/internal/errors"
	"github.com/zitadel/zitadel/internal/eventstore"
	"github.com/zitadel/zitadel/internal/id"
	id_mock "github.com/zitadel/zitadel/internal/id/mock"
	"github.com/zitadel/zitadel/internal/notification/channels/smtp"
	"github.com/zitadel/zitadel/internal/repository/instance"
)

func TestCommandSide_AddSMTPConfig(t *testing.T) {
	type fields struct {
		eventstore  *eventstore.Eventstore
		idGenerator id.Generator
		alg         crypto.EncryptionAlgorithm
	}
	type args struct {
		ctx        context.Context
		instanceID string
		smtp       *smtp.Config
	}
	type res struct {
		want *domain.ObjectDetails
		err  func(error) bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		res    res
	}{
		{
			name: "smtp config, custom domain not existing",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							instance.NewDomainPolicyAddedEvent(context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								true,
								true,
								true,
							),
						),
					),
				),
				idGenerator: id_mock.NewIDGeneratorExpectIDs(t, "configid"),
				alg:         crypto.CreateMockEncryptionAlg(gomock.NewController(t)),
			},
			args: args{
				ctx: authz.WithInstanceID(context.Background(), "INSTANCE"),
				smtp: &smtp.Config{
					SMTP: smtp.SMTP{
						Host:         "host:587",
						User:         "user",
						Password:     "password",
						ProviderType: 1,
					},
					Tls:            true,
					From:           "from@domain.ch",
					FromName:       "name",
					ReplyToAddress: "replyto@domain.ch",
				},
			},
			res: res{
				err: caos_errs.IsErrorInvalidArgument,
			},
		},
		{
			name: "add smtp config, ok",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							instance.NewDomainAddedEvent(context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								"domain.ch",
								false,
							),
						),
						eventFromEventPusher(
							instance.NewDomainPolicyAddedEvent(context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								true, true, false,
							),
						),
					),
					expectPush(
						instance.NewSMTPConfigAddedEvent(
							context.Background(),
							&instance.NewAggregate("INSTANCE").Aggregate,
							"configid",
							true,
							"from@domain.ch",
							"name",
							"",
							"host:587",
							"user",
							&crypto.CryptoValue{
								CryptoType: crypto.TypeEncryption,
								Algorithm:  "enc",
								KeyID:      "id",
								Crypted:    []byte("password"),
							},
							7,
						),
					),
				),
				idGenerator: id_mock.NewIDGeneratorExpectIDs(t, "configid"),
				alg:         crypto.CreateMockEncryptionAlg(gomock.NewController(t)),
			},
			args: args{
				ctx: authz.WithInstanceID(context.Background(), "INSTANCE"),
				smtp: &smtp.Config{
					Tls:      true,
					From:     "from@domain.ch",
					FromName: "name",
					SMTP: smtp.SMTP{
						Host:         "host:587",
						User:         "user",
						Password:     "password",
						ProviderType: 7,
					},
				},
			},
			res: res{
				want: &domain.ObjectDetails{
					ResourceOwner: "INSTANCE",
				},
			},
		},
		{
			name: "add smtp config with reply to address, ok",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							instance.NewDomainAddedEvent(context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								"domain.ch",
								false,
							),
						),
						eventFromEventPusher(
							instance.NewDomainPolicyAddedEvent(context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								true, true, false,
							),
						),
					),
					expectPush(
						instance.NewSMTPConfigAddedEvent(
							context.Background(),
							&instance.NewAggregate("INSTANCE").Aggregate,
							"configid",
							true,
							"from@domain.ch",
							"name",
							"replyto@domain.ch",
							"host:587",
							"user",
							&crypto.CryptoValue{
								CryptoType: crypto.TypeEncryption,
								Algorithm:  "enc",
								KeyID:      "id",
								Crypted:    []byte("password"),
							},
							7,
						),
					),
				),
				idGenerator: id_mock.NewIDGeneratorExpectIDs(t, "configid"),
				alg:         crypto.CreateMockEncryptionAlg(gomock.NewController(t)),
			},
			args: args{
				ctx: authz.WithInstanceID(context.Background(), "INSTANCE"),
				smtp: &smtp.Config{
					Tls:            true,
					From:           "from@domain.ch",
					FromName:       "name",
					ReplyToAddress: "replyto@domain.ch",
					SMTP: smtp.SMTP{
						Host:         "host:587",
						User:         "user",
						Password:     "password",
						ProviderType: 7,
					},
				},
			},
			res: res{
				want: &domain.ObjectDetails{
					ResourceOwner: "INSTANCE",
				},
			},
		},
		{
			name: "smtp config, port is missing",
			fields: fields{
				eventstore:  eventstoreExpect(t),
				idGenerator: id_mock.NewIDGeneratorExpectIDs(t, "configid"),
				alg:         crypto.CreateMockEncryptionAlg(gomock.NewController(t)),
			},
			args: args{
				ctx: authz.WithInstanceID(context.Background(), "INSTANCE"),
				smtp: &smtp.Config{
					Tls:      true,
					From:     "from@domain.ch",
					FromName: "name",
					SMTP: smtp.SMTP{
						Host:     "host",
						User:     "user",
						Password: "password",
					},
				},
			},
			res: res{
				err: caos_errs.IsErrorInvalidArgument,
			},
		},
		{
			name: "smtp config, host is empty",
			fields: fields{
				eventstore:  eventstoreExpect(t),
				idGenerator: id_mock.NewIDGeneratorExpectIDs(t, "configid"),
				alg:         crypto.CreateMockEncryptionAlg(gomock.NewController(t)),
			},
			args: args{
				ctx: authz.WithInstanceID(context.Background(), "INSTANCE"),
				smtp: &smtp.Config{
					Tls:      true,
					From:     "from@domain.ch",
					FromName: "name",
					SMTP: smtp.SMTP{
						Host:         "   ",
						User:         "user",
						Password:     "password",
						ProviderType: 1,
					},
				},
			},
			res: res{
				err: caos_errs.IsErrorInvalidArgument,
			},
		},
		{
			name: "add smtp config, ipv6 works",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							instance.NewDomainAddedEvent(context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								"domain.ch",
								false,
							),
						),
						eventFromEventPusher(
							instance.NewDomainPolicyAddedEvent(context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								true, true, false,
							),
						),
					),
					expectPush(
						instance.NewSMTPConfigAddedEvent(
							context.Background(),
							&instance.NewAggregate("INSTANCE").Aggregate,
							"configid",
							true,
							"from@domain.ch",
							"name",
							"",
							"[2001:db8::1]:2525",
							"user",
							&crypto.CryptoValue{
								CryptoType: crypto.TypeEncryption,
								Algorithm:  "enc",
								KeyID:      "id",
								Crypted:    []byte("password"),
							},
							7,
						),
					),
				),
				alg:         crypto.CreateMockEncryptionAlg(gomock.NewController(t)),
				idGenerator: id_mock.NewIDGeneratorExpectIDs(t, "configid"),
			},
			args: args{
				ctx: authz.WithInstanceID(context.Background(), "INSTANCE"),
				smtp: &smtp.Config{
					Tls:      true,
					From:     "from@domain.ch",
					FromName: "name",
					SMTP: smtp.SMTP{
						Host:         "[2001:db8::1]:2525",
						User:         "user",
						Password:     "password",
						ProviderType: 7,
					},
				},
			},
			res: res{
				want: &domain.ObjectDetails{
					ResourceOwner: "INSTANCE",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Commands{
				eventstore:     tt.fields.eventstore,
				idGenerator:    tt.fields.idGenerator,
				smtpEncryption: tt.fields.alg,
			}
			_, got, err := r.AddSMTPConfig(tt.args.ctx, tt.args.instanceID, tt.args.smtp)
			if tt.res.err == nil {
				assert.NoError(t, err)
			}
			if tt.res.err != nil && !tt.res.err(err) {
				t.Errorf("got wrong err: %v ", err)
			}
			if tt.res.err == nil {
				assert.Equal(t, tt.res.want, got)
			}
		})
	}
}

func TestCommandSide_ChangeSMTPConfig(t *testing.T) {
	type fields struct {
		eventstore *eventstore.Eventstore
	}
	type args struct {
		ctx        context.Context
		instanceID string
		id         string
		smtp       *smtp.Config
	}
	type res struct {
		want *domain.ObjectDetails
		err  func(error) bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		res    res
	}{
		{
			name: "id empty, precondition error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
				),
			},
			args: args{
				ctx:  authz.WithInstanceID(context.Background(), "INSTANCE"),
				smtp: &smtp.Config{},
			},
			res: res{
				err: caos_errs.IsErrorInvalidArgument,
			},
		},
		{
			name: "empty config, invalid argument error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
				),
			},
			args: args{
				ctx:  authz.WithInstanceID(context.Background(), "INSTANCE"),
				smtp: &smtp.Config{},
				id:   "configID",
			},
			res: res{
				err: caos_errs.IsErrorInvalidArgument,
			},
		},
		{
			name: "smtp not existing, not found error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(),
				),
			},
			args: args{
				ctx: authz.WithInstanceID(context.Background(), "INSTANCE"),
				smtp: &smtp.Config{
					Tls:      true,
					From:     "from@domain.ch",
					FromName: "name",
					SMTP: smtp.SMTP{
						Host:         "host:587",
						User:         "user",
						ProviderType: 1,
					},
				},
				instanceID: "INSTANCE",
				id:         "ID",
			},
			res: res{
				err: caos_errs.IsNotFound,
			},
		},
		{
			name: "smtp domain not matched",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							instance.NewDomainAddedEvent(context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								"domain.ch",
								false,
							),
						),
						eventFromEventPusher(
							instance.NewDomainPolicyAddedEvent(context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								true, true, true,
							),
						),
						eventFromEventPusher(
							instance.NewSMTPConfigAddedEvent(
								context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								"ID",
								true,
								"from@domain.ch",
								"name",
								"",
								"host:587",
								"user",
								&crypto.CryptoValue{},
								7,
							),
						),
					),
				),
			},
			args: args{
				ctx:        authz.WithInstanceID(context.Background(), "INSTANCE"),
				instanceID: "INSTANCE",
				id:         "ID",
				smtp: &smtp.Config{
					Tls:      true,
					From:     "from@wrongdomain.ch",
					FromName: "name",
					SMTP: smtp.SMTP{
						Host:         "host:587",
						User:         "user",
						ProviderType: 7,
					},
				},
			},
			res: res{
				err: caos_errs.IsErrorInvalidArgument,
			},
		},
		{
			name: "no changes, precondition error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							instance.NewDomainAddedEvent(context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								"domain.ch",
								false,
							),
						),
						eventFromEventPusher(
							instance.NewDomainPolicyAddedEvent(context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								true, true, true,
							),
						),
						eventFromEventPusher(
							instance.NewSMTPConfigAddedEvent(
								context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								"ID",
								true,
								"from@domain.ch",
								"name",
								"",
								"host:587",
								"user",
								&crypto.CryptoValue{},
								7,
							),
						),
					),
				),
			},
			args: args{
				ctx: authz.WithInstanceID(context.Background(), "INSTANCE"),
				smtp: &smtp.Config{
					Tls:      true,
					From:     "from@domain.ch",
					FromName: "name",
					SMTP: smtp.SMTP{
						Host:         "host:587",
						User:         "user",
						ProviderType: 7,
					},
				},
				instanceID: "INSTANCE",
				id:         "ID",
			},
			res: res{
				err: caos_errs.IsPreconditionFailed,
			},
		},
		{
			name: "smtp config change, ok",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							instance.NewDomainAddedEvent(context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								"domain.ch",
								false,
							),
						),
						eventFromEventPusher(
							instance.NewDomainPolicyAddedEvent(context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								true, true, true,
							),
						),
						eventFromEventPusher(
							instance.NewSMTPConfigAddedEvent(
								context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								"ID",
								true,
								"from@domain.ch",
								"name",
								"",
								"host:587",
								"user",
								&crypto.CryptoValue{},
								1,
							),
						),
					),
					expectPush(
						newSMTPConfigChangedEvent(
							context.Background(),
							"ID",
							false,
							"from2@domain.ch",
							"name2",
							"replyto@domain.ch",
							"host2:587",
							"user2",
							7,
						),
					),
				),
			},
			args: args{
				ctx: authz.WithInstanceID(context.Background(), "INSTANCE"),
				smtp: &smtp.Config{
					Tls:            false,
					From:           "from2@domain.ch",
					FromName:       "name2",
					ReplyToAddress: "replyto@domain.ch",
					SMTP: smtp.SMTP{
						Host:         "host2:587",
						User:         "user2",
						ProviderType: 7,
					},
				},
				id:         "ID",
				instanceID: "INSTANCE",
			},
			res: res{
				want: &domain.ObjectDetails{
					ResourceOwner: "INSTANCE",
				},
			},
		},
		{
			name: "smtp config, port is missing",
			fields: fields{
				eventstore: eventstoreExpect(t),
			},
			args: args{
				ctx: authz.WithInstanceID(context.Background(), "INSTANCE"),
				smtp: &smtp.Config{
					Tls:      true,
					From:     "from@domain.ch",
					FromName: "name",
					SMTP: smtp.SMTP{
						Host:         "host",
						User:         "user",
						Password:     "password",
						ProviderType: 1,
					},
				},
				instanceID: "INSTANCE",
				id:         "ID",
			},
			res: res{
				err: caos_errs.IsErrorInvalidArgument,
			},
		},
		{
			name: "smtp config, host is empty",
			fields: fields{
				eventstore: eventstoreExpect(t),
			},
			args: args{
				ctx: authz.WithInstanceID(context.Background(), "INSTANCE"),
				smtp: &smtp.Config{
					Tls:      true,
					From:     "from@domain.ch",
					FromName: "name",
					SMTP: smtp.SMTP{
						Host:         "   ",
						User:         "user",
						Password:     "password",
						ProviderType: 1,
					},
				},
				instanceID: "INSTANCE",
				id:         "ID",
			},
			res: res{
				err: caos_errs.IsErrorInvalidArgument,
			},
		},
		{
			name: "smtp config change, ipv6 works",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							instance.NewDomainAddedEvent(context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								"domain.ch",
								false,
							),
						),
						eventFromEventPusher(
							instance.NewDomainPolicyAddedEvent(context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								true, true, true,
							),
						),
						eventFromEventPusher(
							instance.NewSMTPConfigAddedEvent(
								context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								"ID",
								true,
								"from@domain.ch",
								"name",
								"",
								"host:587",
								"user",
								&crypto.CryptoValue{},
								7,
							),
						),
					),
					expectPush(
						newSMTPConfigChangedEvent(
							context.Background(),
							"ID",
							false,
							"from2@domain.ch",
							"name2",
							"replyto@domain.ch",
							"[2001:db8::1]:2525",
							"user2",
							2,
						),
					),
				),
			},
			args: args{
				ctx: authz.WithInstanceID(context.Background(), "INSTANCE"),
				smtp: &smtp.Config{
					Tls:            false,
					From:           "from2@domain.ch",
					FromName:       "name2",
					ReplyToAddress: "replyto@domain.ch",
					SMTP: smtp.SMTP{
						Host:         "[2001:db8::1]:2525",
						User:         "user2",
						ProviderType: 2,
					},
				},
				instanceID: "INSTANCE",
				id:         "ID",
			},
			res: res{
				want: &domain.ObjectDetails{
					ResourceOwner: "INSTANCE",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Commands{
				eventstore: tt.fields.eventstore,
			}
			got, err := r.ChangeSMTPConfig(tt.args.ctx, tt.args.instanceID, tt.args.id, tt.args.smtp)
			if tt.res.err == nil {
				assert.NoError(t, err)
			}
			if tt.res.err != nil && !tt.res.err(err) {
				t.Errorf("got wrong err: %v ", err)
			}
			if tt.res.err == nil {
				assert.Equal(t, tt.res.want, got)
			}
		})
	}
}

func TestCommandSide_ChangeSMTPConfigPassword(t *testing.T) {
	type fields struct {
		eventstore *eventstore.Eventstore
		alg        crypto.EncryptionAlgorithm
	}
	type args struct {
		ctx        context.Context
		instanceID string
		id         string
		password   string
	}
	type res struct {
		want *domain.ObjectDetails
		err  func(error) bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		res    res
	}{
		{
			name: "smtp config, error not found",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(),
				),
			},
			args: args{
				ctx:      context.Background(),
				password: "",
				id:       "ID",
			},
			res: res{
				err: caos_errs.IsNotFound,
			},
		},
		{
			name: "change smtp config password, ok",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							instance.NewSMTPConfigAddedEvent(
								context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								"ID",
								true,
								"from",
								"name",
								"",
								"host:587",
								"user",
								&crypto.CryptoValue{},
								7,
							),
						),
					),
					expectPush(
						instance.NewSMTPConfigPasswordChangedEvent(
							context.Background(),
							&instance.NewAggregate("INSTANCE").Aggregate,
							"ID",
							&crypto.CryptoValue{
								CryptoType: crypto.TypeEncryption,
								Algorithm:  "enc",
								KeyID:      "id",
								Crypted:    []byte("password"),
							},
						),
					),
				),
				alg: crypto.CreateMockEncryptionAlg(gomock.NewController(t)),
			},
			args: args{
				ctx:        authz.WithInstanceID(context.Background(), "INSTANCE"),
				password:   "password",
				id:         "ID",
				instanceID: "INSTANCE",
			},
			res: res{
				want: &domain.ObjectDetails{
					ResourceOwner: "INSTANCE",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Commands{
				eventstore:     tt.fields.eventstore,
				smtpEncryption: tt.fields.alg,
			}
			got, err := r.ChangeSMTPConfigPassword(tt.args.ctx, tt.args.instanceID, tt.args.id, tt.args.password)
			if tt.res.err == nil {
				assert.NoError(t, err)
			}
			if tt.res.err != nil && !tt.res.err(err) {
				t.Errorf("got wrong err: %v ", err)
			}
			if tt.res.err == nil {
				assert.Equal(t, tt.res.want, got)
			}
		})
	}
}

func TestCommandSide_ActivateSMTPConfig(t *testing.T) {
	type fields struct {
		eventstore *eventstore.Eventstore
		alg        crypto.EncryptionAlgorithm
	}
	type args struct {
		ctx         context.Context
		instanceID  string
		id          string
		activatedId string
	}
	type res struct {
		want *domain.ObjectDetails
		err  func(error) bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		res    res
	}{
		{
			name: "id empty, precondition error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
				),
			},
			args: args{
				ctx: authz.WithInstanceID(context.Background(), "INSTANCE"),
			},
			res: res{
				err: caos_errs.IsErrorInvalidArgument,
			},
		},
		{
			name: "smtp not existing, not found error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(),
				),
			},
			args: args{
				ctx:        context.Background(),
				instanceID: "INSTANCE",
				id:         "id",
			},
			res: res{
				err: caos_errs.IsNotFound,
			},
		},
		{
			name: "activate smtp config, ok",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							instance.NewSMTPConfigAddedEvent(
								context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								"ID",
								true,
								"from",
								"name",
								"",
								"host:587",
								"user",
								&crypto.CryptoValue{},
								7,
							),
						),
					),
					expectPush(
						instance.NewSMTPConfigActivatedEvent(
							context.Background(),
							&instance.NewAggregate("INSTANCE").Aggregate,
							"ID",
						),
					),
				),
			},
			args: args{
				ctx:         authz.WithInstanceID(context.Background(), "INSTANCE"),
				id:          "ID",
				instanceID:  "INSTANCE",
				activatedId: "",
			},
			res: res{
				want: &domain.ObjectDetails{
					ResourceOwner: "INSTANCE",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Commands{
				eventstore:     tt.fields.eventstore,
				smtpEncryption: tt.fields.alg,
			}
			got, err := r.ActivateSMTPConfig(tt.args.ctx, tt.args.instanceID, tt.args.id, tt.args.activatedId)
			if tt.res.err == nil {
				assert.NoError(t, err)
			}
			if tt.res.err != nil && !tt.res.err(err) {
				t.Errorf("got wrong err: %v ", err)
			}
			if tt.res.err == nil {
				assert.Equal(t, tt.res.want, got)
			}
		})
	}
}

func TestCommandSide_DeactivateSMTPConfig(t *testing.T) {
	type fields struct {
		eventstore *eventstore.Eventstore
		alg        crypto.EncryptionAlgorithm
	}
	type args struct {
		ctx         context.Context
		instanceID  string
		id          string
		activatedId string
	}
	type res struct {
		want *domain.ObjectDetails
		err  func(error) bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		res    res
	}{
		{
			name: "id empty, precondition error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
				),
			},
			args: args{
				ctx: authz.WithInstanceID(context.Background(), "INSTANCE"),
			},
			res: res{
				err: caos_errs.IsErrorInvalidArgument,
			},
		},
		{
			name: "smtp not existing, not found error",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(),
				),
			},
			args: args{
				ctx:        context.Background(),
				instanceID: "INSTANCE",
				id:         "id",
			},
			res: res{
				err: caos_errs.IsNotFound,
			},
		},
		{
			name: "deactivate smtp config, ok",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							instance.NewSMTPConfigAddedEvent(
								context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								"ID",
								true,
								"from",
								"name",
								"",
								"host:587",
								"user",
								&crypto.CryptoValue{},
								7,
							),
						),
						eventFromEventPusher(
							instance.NewSMTPConfigActivatedEvent(
								context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								"ID",
							),
						),
					),
					expectPush(
						instance.NewSMTPConfigDeactivatedEvent(
							context.Background(),
							&instance.NewAggregate("INSTANCE").Aggregate,
							"ID",
						),
					),
				),
			},
			args: args{
				ctx:         authz.WithInstanceID(context.Background(), "INSTANCE"),
				id:          "ID",
				instanceID:  "INSTANCE",
				activatedId: "",
			},
			res: res{
				want: &domain.ObjectDetails{
					ResourceOwner: "INSTANCE",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Commands{
				eventstore:     tt.fields.eventstore,
				smtpEncryption: tt.fields.alg,
			}
			got, err := r.DeactivateSMTPConfig(tt.args.ctx, tt.args.instanceID, tt.args.id)
			if tt.res.err == nil {
				assert.NoError(t, err)
			}
			if tt.res.err != nil && !tt.res.err(err) {
				t.Errorf("got wrong err: %v ", err)
			}
			if tt.res.err == nil {
				assert.Equal(t, tt.res.want, got)
			}
		})
	}
}

func TestCommandSide_RemoveSMTPConfig(t *testing.T) {
	type fields struct {
		eventstore *eventstore.Eventstore
		alg        crypto.EncryptionAlgorithm
	}
	type args struct {
		ctx        context.Context
		instanceID string
		id         string
	}
	type res struct {
		want *domain.ObjectDetails
		err  func(error) bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		res    res
	}{
		{
			name: "smtp config, error not found",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(),
				),
			},
			args: args{
				ctx: context.Background(),
				id:  "ID",
			},
			res: res{
				err: caos_errs.IsNotFound,
			},
		},
		{
			name: "remove smtp config, ok",
			fields: fields{
				eventstore: eventstoreExpect(
					t,
					expectFilter(
						eventFromEventPusher(
							instance.NewSMTPConfigAddedEvent(
								context.Background(),
								&instance.NewAggregate("INSTANCE").Aggregate,
								"ID",
								true,
								"from",
								"name",
								"",
								"host:587",
								"user",
								&crypto.CryptoValue{},
								7,
							),
						),
					),
					expectPush(
						instance.NewSMTPConfigRemovedEvent(
							context.Background(),
							&instance.NewAggregate("INSTANCE").Aggregate,
							"ID",
						),
					),
				),
			},
			args: args{
				ctx:        authz.WithInstanceID(context.Background(), "INSTANCE"),
				id:         "ID",
				instanceID: "INSTANCE",
			},
			res: res{
				want: &domain.ObjectDetails{
					ResourceOwner: "INSTANCE",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Commands{
				eventstore:     tt.fields.eventstore,
				smtpEncryption: tt.fields.alg,
			}
			got, err := r.RemoveSMTPConfig(tt.args.ctx, tt.args.instanceID, tt.args.id)
			if tt.res.err == nil {
				assert.NoError(t, err)
			}
			if tt.res.err != nil && !tt.res.err(err) {
				t.Errorf("got wrong err: %v ", err)
			}
			if tt.res.err == nil {
				assert.Equal(t, tt.res.want, got)
			}
		})
	}
}

func newSMTPConfigChangedEvent(ctx context.Context, id string, tls bool, fromAddress, fromName, replyTo, host, user string, providerType uint32) *instance.SMTPConfigChangedEvent {
	changes := []instance.SMTPConfigChanges{
		instance.ChangeSMTPConfigTLS(tls),
		instance.ChangeSMTPConfigFromAddress(fromAddress),
		instance.ChangeSMTPConfigFromName(fromName),
		instance.ChangeSMTPConfigReplyToAddress(replyTo),
		instance.ChangeSMTPConfigSMTPHost(host),
		instance.ChangeSMTPConfigSMTPUser(user),
		instance.ChangeSMTPConfigProviderType(providerType),
	}
	event, _ := instance.NewSMTPConfigChangeEvent(ctx,
		&instance.NewAggregate("INSTANCE").Aggregate,
		id,
		changes,
	)
	return event
}
