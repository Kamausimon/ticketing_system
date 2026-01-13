--
-- PostgreSQL database dump
--

\restrict euhg8NAxHPDbyzdM1rc8sbuidy5sW9ShGZV2Jgw7fd0BkaQYauvZjJYY55RcxwN

-- Dumped from database version 16.11 (Ubuntu 16.11-0ubuntu0.24.04.1)
-- Dumped by pg_dump version 16.11 (Ubuntu 16.11-0ubuntu0.24.04.1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: account_activities; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.account_activities (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    account_id bigint NOT NULL,
    user_id bigint,
    action character varying(100) NOT NULL,
    category character varying(50),
    description text NOT NULL,
    ip_address character varying(45),
    user_agent text,
    success boolean DEFAULT true,
    metadata jsonb,
    severity character varying(20) DEFAULT 'info'::character varying,
    resource character varying(100),
    resource_id bigint,
    "timestamp" timestamp with time zone NOT NULL
);


ALTER TABLE public.account_activities OWNER TO postgres;

--
-- Name: account_activities_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.account_activities_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.account_activities_id_seq OWNER TO postgres;

--
-- Name: account_activities_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.account_activities_id_seq OWNED BY public.account_activities.id;


--
-- Name: account_payment_gateways; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.account_payment_gateways (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    account_id bigint NOT NULL,
    payment_gateway_id bigint NOT NULL,
    config text
);


ALTER TABLE public.account_payment_gateways OWNER TO postgres;

--
-- Name: account_payment_gateways_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.account_payment_gateways_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.account_payment_gateways_id_seq OWNER TO postgres;

--
-- Name: account_payment_gateways_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.account_payment_gateways_id_seq OWNED BY public.account_payment_gateways.id;


--
-- Name: accounts; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.accounts (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    first_name text,
    last_name text,
    email text NOT NULL,
    timezone_id bigint,
    date_format_id bigint,
    date_time_format_id bigint,
    currency_id bigint,
    last_ip text,
    last_login_date timestamp with time zone,
    address1 text,
    address2 text,
    city text,
    county text,
    postal_code text,
    is_active boolean DEFAULT true,
    is_banned boolean DEFAULT false,
    stripe_access_token text,
    stripe_refresh_token text,
    stripe_secret_key text,
    stripe_publishable_key text,
    stripe_data_raw text,
    payment_gateway_id bigint
);


ALTER TABLE public.accounts OWNER TO postgres;

--
-- Name: accounts_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.accounts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.accounts_id_seq OWNER TO postgres;

--
-- Name: accounts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.accounts_id_seq OWNED BY public.accounts.id;


--
-- Name: attendees; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.attendees (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    order_id bigint NOT NULL,
    event_id bigint NOT NULL,
    ticket_id bigint NOT NULL,
    first_name text,
    last_name text,
    email text,
    has_arrived boolean,
    arrival_time timestamp with time zone,
    account_id bigint NOT NULL,
    is_refunded boolean DEFAULT false,
    private_reference_number bigint
);


ALTER TABLE public.attendees OWNER TO postgres;

--
-- Name: attendees_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.attendees_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.attendees_id_seq OWNER TO postgres;

--
-- Name: attendees_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.attendees_id_seq OWNED BY public.attendees.id;


--
-- Name: currencies; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.currencies (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    code character varying(3) NOT NULL,
    name text NOT NULL,
    symbol text NOT NULL,
    is_active boolean DEFAULT true
);


ALTER TABLE public.currencies OWNER TO postgres;

--
-- Name: currencies_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.currencies_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.currencies_id_seq OWNER TO postgres;

--
-- Name: currencies_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.currencies_id_seq OWNED BY public.currencies.id;


--
-- Name: date_formats; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.date_formats (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    format text NOT NULL,
    example text NOT NULL,
    is_active boolean DEFAULT true
);


ALTER TABLE public.date_formats OWNER TO postgres;

--
-- Name: date_formats_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.date_formats_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.date_formats_id_seq OWNER TO postgres;

--
-- Name: date_formats_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.date_formats_id_seq OWNED BY public.date_formats.id;


--
-- Name: date_time_formats; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.date_time_formats (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    format text NOT NULL,
    example text NOT NULL,
    is_active boolean DEFAULT true
);


ALTER TABLE public.date_time_formats OWNER TO postgres;

--
-- Name: date_time_formats_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.date_time_formats_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.date_time_formats_id_seq OWNER TO postgres;

--
-- Name: date_time_formats_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.date_time_formats_id_seq OWNED BY public.date_time_formats.id;


--
-- Name: email_verifications; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.email_verifications (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id bigint,
    token text,
    email text,
    status text DEFAULT 'pending'::text,
    verified_at timestamp with time zone,
    expires_at timestamp with time zone,
    last_sent_at timestamp with time zone,
    resend_count bigint DEFAULT 0,
    max_resends bigint DEFAULT 3,
    ip_address text,
    user_agent text,
    issued_at timestamp with time zone
);


ALTER TABLE public.email_verifications OWNER TO postgres;

--
-- Name: email_verifications_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.email_verifications_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.email_verifications_id_seq OWNER TO postgres;

--
-- Name: email_verifications_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.email_verifications_id_seq OWNED BY public.email_verifications.id;


--
-- Name: event_images; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.event_images (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    image_path text,
    event_id bigint NOT NULL,
    account_id bigint NOT NULL,
    user_id bigint NOT NULL
);


ALTER TABLE public.event_images OWNER TO postgres;

--
-- Name: event_images_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.event_images_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.event_images_id_seq OWNER TO postgres;

--
-- Name: event_images_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.event_images_id_seq OWNED BY public.event_images.id;


--
-- Name: event_metrics; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.event_metrics (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    event_id bigint NOT NULL,
    date date NOT NULL,
    hour bigint,
    page_views bigint DEFAULT 0,
    unique_visitors bigint DEFAULT 0,
    bounce_rate numeric DEFAULT 0,
    avg_time_on_page bigint DEFAULT 0,
    add_to_cart bigint DEFAULT 0,
    checkout_start bigint DEFAULT 0,
    checkout_complete bigint DEFAULT 0,
    conversion_rate numeric DEFAULT 0,
    gross_revenue bigint DEFAULT 0,
    net_revenue bigint DEFAULT 0,
    platform_fees bigint DEFAULT 0,
    payment_fees bigint DEFAULT 0,
    refund_amount bigint DEFAULT 0,
    tickets_sold bigint DEFAULT 0,
    tickets_refunded bigint DEFAULT 0,
    tickets_checked_in bigint DEFAULT 0,
    inventory_remaining bigint DEFAULT 0,
    promo_code_uses bigint DEFAULT 0,
    promo_discount bigint DEFAULT 0,
    top_countries text,
    top_cities text,
    mobile_percent numeric DEFAULT 0,
    desktop_percent numeric DEFAULT 0,
    app_percent numeric DEFAULT 0,
    CONSTRAINT chk_event_metrics_hour CHECK (((hour >= 0) AND (hour <= 23)))
);


ALTER TABLE public.event_metrics OWNER TO postgres;

--
-- Name: event_metrics_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.event_metrics_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.event_metrics_id_seq OWNER TO postgres;

--
-- Name: event_metrics_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.event_metrics_id_seq OWNED BY public.event_metrics.id;


--
-- Name: event_stats; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.event_stats (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    date timestamp with time zone,
    day date,
    hour bigint,
    views bigint,
    unique_views bigint,
    tickets_sold bigint,
    sales_volume numeric,
    organizer_fee_volume numeric,
    event_id bigint NOT NULL,
    add_to_cart_count bigint,
    check_out_start bigint,
    conversion_rate numeric,
    gross_revenue numeric,
    net_revenue numeric,
    platform_fees numeric,
    payment_fees numeric,
    average_time_on_page bigint,
    bounce_rate numeric,
    granularity text NOT NULL,
    CONSTRAINT chk_event_stats_hour CHECK (((hour >= 0) AND (hour <= 23)))
);


ALTER TABLE public.event_stats OWNER TO postgres;

--
-- Name: event_stats_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.event_stats_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.event_stats_id_seq OWNER TO postgres;

--
-- Name: event_stats_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.event_stats_id_seq OWNED BY public.event_stats.id;


--
-- Name: event_venues; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.event_venues (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    venue_id bigint NOT NULL,
    event_id bigint NOT NULL,
    venue_role text DEFAULT 'primary'::text,
    setup_time timestamp with time zone,
    event_time timestamp with time zone,
    cleanup_time timestamp with time zone
);


ALTER TABLE public.event_venues OWNER TO postgres;

--
-- Name: event_venues_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.event_venues_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.event_venues_id_seq OWNER TO postgres;

--
-- Name: event_venues_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.event_venues_id_seq OWNED BY public.event_venues.id;


--
-- Name: events; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.events (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    title text NOT NULL,
    location text NOT NULL,
    bg_type text,
    bg_color text,
    description text NOT NULL,
    start_date timestamp with time zone NOT NULL,
    end_date timestamp with time zone NOT NULL,
    on_sale_date timestamp with time zone,
    organizer_id bigint NOT NULL,
    account_id bigint NOT NULL,
    sales_volume numeric,
    organizer_fees_volume numeric,
    organizer_fee_fixed numeric,
    organizer_fee_percentage numeric,
    currency text NOT NULL,
    location_address text,
    location_address_line text,
    location_country text,
    pre_order_message_display text,
    post_order_message_display text,
    is_live boolean DEFAULT false NOT NULL,
    barcode_type text NOT NULL,
    is_barcode_enabled boolean DEFAULT false NOT NULL,
    ticket_border_color text,
    ticket_bg_color text,
    ticket_text_color text,
    ticket_sub_text_color text,
    enable_offline_payment boolean DEFAULT false NOT NULL,
    max_capacity bigint,
    status text DEFAULT 'draft'::text NOT NULL,
    category text NOT NULL,
    tags text,
    min_age bigint,
    is_private boolean DEFAULT false
);


ALTER TABLE public.events OWNER TO postgres;

--
-- Name: events_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.events_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.events_id_seq OWNER TO postgres;

--
-- Name: events_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.events_id_seq OWNED BY public.events.id;


--
-- Name: login_history; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.login_history (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    account_id bigint NOT NULL,
    user_id bigint,
    ip_address character varying(45) NOT NULL,
    user_agent text,
    location character varying(255),
    device character varying(100),
    browser character varying(100),
    success boolean NOT NULL,
    fail_reason character varying(255),
    login_at timestamp with time zone NOT NULL,
    logout_at timestamp with time zone,
    session_duration bigint
);


ALTER TABLE public.login_history OWNER TO postgres;

--
-- Name: login_history_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.login_history_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.login_history_id_seq OWNER TO postgres;

--
-- Name: login_history_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.login_history_id_seq OWNED BY public.login_history.id;


--
-- Name: notification_preferences; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.notification_preferences (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    account_id bigint NOT NULL,
    email_notifications boolean DEFAULT true,
    sms_notifications boolean DEFAULT false,
    push_notifications boolean DEFAULT true,
    event_updates boolean DEFAULT true,
    payment_notifications boolean DEFAULT true,
    security_alerts boolean DEFAULT true,
    marketing_emails boolean DEFAULT false
);


ALTER TABLE public.notification_preferences OWNER TO postgres;

--
-- Name: notification_preferences_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.notification_preferences_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.notification_preferences_id_seq OWNER TO postgres;

--
-- Name: notification_preferences_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.notification_preferences_id_seq OWNED BY public.notification_preferences.id;


--
-- Name: order_items; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.order_items (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    order_id bigint NOT NULL,
    ticket_class_id bigint NOT NULL,
    quantity bigint NOT NULL,
    unit_price bigint NOT NULL,
    total_price bigint NOT NULL,
    discount bigint,
    promo_code_used text
);


ALTER TABLE public.order_items OWNER TO postgres;

--
-- Name: order_items_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.order_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.order_items_id_seq OWNER TO postgres;

--
-- Name: order_items_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.order_items_id_seq OWNED BY public.order_items.id;


--
-- Name: orders; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.orders (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    account_id bigint NOT NULL,
    first_name text,
    last_name text,
    email text,
    business_name text,
    business_tax_number text,
    business_address_line text,
    ticket_pdf_path text,
    order_preference text,
    transaction_id bigint,
    discount numeric,
    booking_fee numeric,
    organizer_booking_fee numeric,
    order_date timestamp with time zone,
    notes text,
    is_deleted boolean DEFAULT false,
    is_cancelled boolean DEFAULT false,
    is_partially_refunded boolean DEFAULT false,
    amount numeric,
    amount_refunded numeric,
    event_id bigint NOT NULL,
    payment_gateway_id bigint,
    is_payment_received boolean DEFAULT false,
    is_business boolean,
    tax_amount numeric,
    status text DEFAULT 'pending'::text NOT NULL,
    payment_status text DEFAULT 'pending'::text NOT NULL,
    total_amount bigint NOT NULL,
    currency text DEFAULT 'KSH'::text NOT NULL,
    completed_at timestamp with time zone,
    cancelled_at timestamp with time zone,
    refunded_at timestamp with time zone
);


ALTER TABLE public.orders OWNER TO postgres;

--
-- Name: orders_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.orders_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.orders_id_seq OWNER TO postgres;

--
-- Name: orders_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.orders_id_seq OWNED BY public.orders.id;


--
-- Name: organizers; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.organizers (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    account_id bigint NOT NULL,
    name text,
    about text,
    email text,
    phone text,
    confirmation_key text,
    facebook text,
    twitter text,
    logo_path text,
    is_email_confirmed boolean DEFAULT false,
    show_twitter_widget boolean,
    show_facebook_widget boolean,
    tax_name text,
    tax_value numeric,
    tax_pin text,
    charge_tax bigint,
    page_header_bg_color text,
    page_bg_color text,
    page_text_color text,
    enable_organizer_page boolean,
    payment_gateway_id bigint,
    bank_account_name text,
    bank_account_number text,
    bank_code text,
    bank_country text,
    is_payment_configured boolean DEFAULT false,
    is_verified boolean DEFAULT false,
    verification_status character varying(50) DEFAULT 'pending'::character varying,
    rejection_reason text,
    kyc_status character varying(50) DEFAULT 'pending'::character varying,
    kyc_notes text,
    kyc_completed_at text
);


ALTER TABLE public.organizers OWNER TO postgres;

--
-- Name: organizers_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.organizers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.organizers_id_seq OWNER TO postgres;

--
-- Name: organizers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.organizers_id_seq OWNED BY public.organizers.id;


--
-- Name: password_reset_attempts; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.password_reset_attempts (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    password_reset_id bigint NOT NULL,
    ip_address text NOT NULL,
    user_agent text NOT NULL,
    attempted_at timestamp with time zone NOT NULL,
    was_successful boolean DEFAULT false NOT NULL,
    token_valid boolean DEFAULT false,
    not_expired boolean DEFAULT false,
    ip_matched boolean DEFAULT true,
    rate_limit_passed boolean DEFAULT true,
    failure_reason text,
    error_code text,
    country text,
    city text,
    isp text,
    response_time_ms bigint
);


ALTER TABLE public.password_reset_attempts OWNER TO postgres;

--
-- Name: password_reset_attempts_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.password_reset_attempts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.password_reset_attempts_id_seq OWNER TO postgres;

--
-- Name: password_reset_attempts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.password_reset_attempts_id_seq OWNED BY public.password_reset_attempts.id;


--
-- Name: password_resets; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.password_resets (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    token text NOT NULL,
    email text NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    method text DEFAULT 'email'::text NOT NULL,
    user_id bigint,
    account_id bigint,
    ip_address text NOT NULL,
    user_agent text NOT NULL,
    attempt_count bigint DEFAULT 0,
    max_attempts bigint DEFAULT 3 NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    issued_at timestamp with time zone NOT NULL,
    used_at timestamp with time zone,
    revoked_at timestamp with time zone,
    last_attempt_at timestamp with time zone,
    original_ip text NOT NULL,
    used_from_ip text,
    same_ip_required boolean DEFAULT false,
    require_current_password boolean DEFAULT false,
    require_two_factor boolean DEFAULT false,
    is_security_reset boolean DEFAULT false,
    requested_by bigint,
    approved_by bigint,
    rate_limit_key text,
    previous_reset_at timestamp with time zone,
    cooldown_until timestamp with time zone,
    reset_reason text,
    admin_notes text,
    user_message text,
    should_cleanup boolean DEFAULT true,
    cleanup_after timestamp with time zone NOT NULL
);


ALTER TABLE public.password_resets OWNER TO postgres;

--
-- Name: password_resets_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.password_resets_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.password_resets_id_seq OWNER TO postgres;

--
-- Name: password_resets_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.password_resets_id_seq OWNED BY public.password_resets.id;


--
-- Name: payment_gateways; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.payment_gateways (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    provider_name text,
    provider_url text,
    is_on_site boolean,
    can_refund boolean,
    name text
);


ALTER TABLE public.payment_gateways OWNER TO postgres;

--
-- Name: payment_gateways_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.payment_gateways_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.payment_gateways_id_seq OWNER TO postgres;

--
-- Name: payment_gateways_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.payment_gateways_id_seq OWNED BY public.payment_gateways.id;


--
-- Name: payment_methods; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.payment_methods (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    account_id bigint NOT NULL,
    type text NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    display_name text NOT NULL,
    nickname text,
    is_default boolean DEFAULT false,
    card_brand text,
    card_last4 character varying(4),
    card_expiry_month bigint,
    card_expiry_year bigint,
    card_country character varying(2),
    card_fingerprint text,
    mpesa_phone_number text,
    mpesa_account_name text,
    bank_account_last4 text,
    bank_name text,
    bank_code text,
    bank_account_holder text,
    stripe_payment_method_id text,
    stripe_customer_id text,
    external_payment_method_id text,
    is_verified boolean DEFAULT false,
    verified_at timestamp with time zone,
    last_used_at timestamp with time zone,
    failure_count bigint DEFAULT 0,
    last_failure_at timestamp with time zone,
    billing_address text,
    metadata text
);


ALTER TABLE public.payment_methods OWNER TO postgres;

--
-- Name: payment_methods_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.payment_methods_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.payment_methods_id_seq OWNER TO postgres;

--
-- Name: payment_methods_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.payment_methods_id_seq OWNED BY public.payment_methods.id;


--
-- Name: payment_records; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.payment_records (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    amount bigint NOT NULL,
    currency text DEFAULT 'KSH'::text NOT NULL,
    type text NOT NULL,
    status text NOT NULL,
    order_id bigint,
    event_id bigint,
    account_id bigint,
    organizer_id bigint,
    payment_gateway_id bigint,
    external_transaction_id text,
    external_reference text,
    gateway_response_code text,
    initiated_at timestamp with time zone NOT NULL,
    processed_at timestamp with time zone,
    completed_at timestamp with time zone,
    failed_at timestamp with time zone,
    description text NOT NULL,
    notes text,
    ip_address text,
    user_agent text,
    platform_fee_amount bigint DEFAULT 0,
    gateway_fee_amount bigint DEFAULT 0,
    net_amount bigint NOT NULL,
    parent_record_id bigint,
    reconciled_at timestamp with time zone,
    reconciliation_ref text
);


ALTER TABLE public.payment_records OWNER TO postgres;

--
-- Name: payment_records_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.payment_records_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.payment_records_id_seq OWNER TO postgres;

--
-- Name: payment_records_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.payment_records_id_seq OWNED BY public.payment_records.id;


--
-- Name: payment_transactions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.payment_transactions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    amount bigint NOT NULL,
    currency text DEFAULT 'KSH'::text NOT NULL,
    type text NOT NULL,
    status text NOT NULL,
    order_id bigint,
    payment_gateway_id bigint,
    organizer_id bigint,
    external_transaction_id text,
    external_reference text,
    processed_at timestamp with time zone,
    settled_at timestamp with time zone,
    description text,
    notes text,
    parent_transaction_id bigint
);


ALTER TABLE public.payment_transactions OWNER TO postgres;

--
-- Name: payment_transactions_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.payment_transactions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.payment_transactions_id_seq OWNER TO postgres;

--
-- Name: payment_transactions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.payment_transactions_id_seq OWNED BY public.payment_transactions.id;


--
-- Name: payout_accounts; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.payout_accounts (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    organizer_id bigint NOT NULL,
    account_type text NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    display_name text NOT NULL,
    is_default boolean DEFAULT false,
    bank_name text,
    bank_code text,
    bank_branch text,
    bank_country character varying(2),
    account_number text,
    account_holder_name text,
    mobile_provider text,
    mobile_phone_number text,
    mobile_account_name text,
    paypal_email text,
    stripe_account_id text,
    stripe_country text,
    currency text DEFAULT 'KSH'::text NOT NULL,
    is_verified boolean DEFAULT false,
    verified_at timestamp with time zone,
    verified_by bigint,
    verification_notes text,
    document_paths text,
    verification_token text,
    total_payouts_count bigint DEFAULT 0,
    total_payouts_amount bigint DEFAULT 0,
    last_payout_at timestamp with time zone,
    last_payout_amount bigint,
    failed_payouts_count bigint DEFAULT 0,
    last_failure_at timestamp with time zone,
    last_failure_reason text,
    requires_kyc boolean DEFAULT false,
    kyc_status text,
    kyc_completed_at timestamp with time zone,
    is_suspicious_activity boolean DEFAULT false,
    suspicion_reason text,
    reviewed_by bigint,
    reviewed_at timestamp with time zone,
    daily_payout_limit bigint,
    monthly_payout_limit bigint,
    requires_approval boolean DEFAULT false,
    external_account_id text,
    external_metadata text,
    address_line1 text,
    address_line2 text,
    city text,
    state text,
    postal_code text,
    country character varying(2),
    notes text
);


ALTER TABLE public.payout_accounts OWNER TO postgres;

--
-- Name: payout_accounts_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.payout_accounts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.payout_accounts_id_seq OWNER TO postgres;

--
-- Name: payout_accounts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.payout_accounts_id_seq OWNED BY public.payout_accounts.id;


--
-- Name: promotion_rules; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.promotion_rules (
    id bigint NOT NULL,
    promotion_id bigint NOT NULL,
    rule_type text NOT NULL,
    rule_operator text NOT NULL,
    rule_value text NOT NULL,
    error_message text NOT NULL,
    is_active boolean DEFAULT true,
    execution_order integer DEFAULT 0,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.promotion_rules OWNER TO postgres;

--
-- Name: promotion_rules_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.promotion_rules_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.promotion_rules_id_seq OWNER TO postgres;

--
-- Name: promotion_rules_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.promotion_rules_id_seq OWNED BY public.promotion_rules.id;


--
-- Name: promotion_usages; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.promotion_usages (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    promotion_id bigint NOT NULL,
    order_id bigint NOT NULL,
    account_id bigint NOT NULL,
    discount_amount bigint NOT NULL,
    original_amount bigint NOT NULL,
    final_amount bigint NOT NULL,
    used_at timestamp with time zone NOT NULL,
    ip_address text,
    user_agent text,
    validation_time bigint,
    cache_hit boolean DEFAULT false
);


ALTER TABLE public.promotion_usages OWNER TO postgres;

--
-- Name: promotion_usages_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.promotion_usages_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.promotion_usages_id_seq OWNER TO postgres;

--
-- Name: promotion_usages_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.promotion_usages_id_seq OWNED BY public.promotion_usages.id;


--
-- Name: promotions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.promotions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    code text NOT NULL,
    name text NOT NULL,
    description text NOT NULL,
    type text NOT NULL,
    status text NOT NULL,
    target text NOT NULL,
    discount_percentage integer,
    discount_amount bigint,
    free_quantity integer,
    minimum_purchase bigint,
    maximum_discount bigint,
    event_id bigint,
    ticket_class_ids text,
    event_categories text,
    organizer_id bigint,
    start_date timestamp with time zone NOT NULL,
    end_date timestamp with time zone NOT NULL,
    early_bird_cutoff timestamp with time zone,
    usage_limit integer,
    usage_count integer DEFAULT 0,
    per_user_limit integer,
    per_order_limit integer DEFAULT 1,
    is_unlimited boolean DEFAULT false,
    precomputed_active boolean DEFAULT false,
    last_usage_check timestamp with time zone,
    first_time_customers boolean DEFAULT false,
    minimum_age integer,
    allowed_user_ids text,
    excluded_user_ids text,
    created_by bigint NOT NULL,
    is_public boolean DEFAULT true,
    requires_approval boolean DEFAULT false,
    total_revenue bigint DEFAULT 0,
    total_discount bigint DEFAULT 0,
    conversion_rate numeric,
    internal_notes text,
    marketing_tags text,
    CONSTRAINT chk_promotions_discount_percentage CHECK ((discount_percentage <= 100))
);


ALTER TABLE public.promotions OWNER TO postgres;

--
-- Name: promotions_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.promotions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.promotions_id_seq OWNER TO postgres;

--
-- Name: promotions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.promotions_id_seq OWNED BY public.promotions.id;


--
-- Name: recovery_codes; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.recovery_codes (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    two_factor_auth_id bigint NOT NULL,
    code_hash character varying(255) NOT NULL,
    used boolean DEFAULT false,
    used_at timestamp with time zone,
    used_from_ip character varying(45)
);


ALTER TABLE public.recovery_codes OWNER TO postgres;

--
-- Name: recovery_codes_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.recovery_codes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.recovery_codes_id_seq OWNER TO postgres;

--
-- Name: recovery_codes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.recovery_codes_id_seq OWNED BY public.recovery_codes.id;


--
-- Name: refund_line_items; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.refund_line_items (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    refund_record_id bigint NOT NULL,
    order_item_id bigint NOT NULL,
    ticket_id bigint,
    quantity bigint NOT NULL,
    refund_amount bigint NOT NULL,
    reason text,
    description text NOT NULL
);


ALTER TABLE public.refund_line_items OWNER TO postgres;

--
-- Name: refund_line_items_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.refund_line_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.refund_line_items_id_seq OWNER TO postgres;

--
-- Name: refund_line_items_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.refund_line_items_id_seq OWNED BY public.refund_line_items.id;


--
-- Name: refund_records; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.refund_records (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    refund_number text NOT NULL,
    refund_type text NOT NULL,
    refund_reason text NOT NULL,
    status text NOT NULL,
    order_id bigint NOT NULL,
    event_id bigint NOT NULL,
    account_id bigint NOT NULL,
    organizer_id bigint NOT NULL,
    original_amount bigint NOT NULL,
    refund_amount bigint NOT NULL,
    organizer_impact bigint NOT NULL,
    currency text DEFAULT 'KSH'::text NOT NULL,
    payment_gateway_id bigint,
    external_refund_id text,
    requested_by bigint,
    approved_by bigint,
    requested_at timestamp with time zone NOT NULL,
    approved_at timestamp with time zone,
    processed_at timestamp with time zone,
    completed_at timestamp with time zone,
    failed_at timestamp with time zone,
    affects_settlement boolean DEFAULT true,
    settlement_adjusted boolean DEFAULT false,
    description text NOT NULL,
    internal_notes text,
    rejection_reason text
);


ALTER TABLE public.refund_records OWNER TO postgres;

--
-- Name: refund_records_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.refund_records_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.refund_records_id_seq OWNER TO postgres;

--
-- Name: refund_records_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.refund_records_id_seq OWNED BY public.refund_records.id;


--
-- Name: reserved_tickets; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.reserved_tickets (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    ticket_id bigint NOT NULL,
    event_id bigint NOT NULL,
    quantity_reserved bigint,
    expires timestamp with time zone,
    session_id text
);


ALTER TABLE public.reserved_tickets OWNER TO postgres;

--
-- Name: reserved_tickets_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.reserved_tickets_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.reserved_tickets_id_seq OWNER TO postgres;

--
-- Name: reserved_tickets_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.reserved_tickets_id_seq OWNED BY public.reserved_tickets.id;


--
-- Name: reset_configurations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.reset_configurations (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    token_length bigint DEFAULT 32 NOT NULL,
    token_expiry_minutes bigint DEFAULT 15 NOT NULL,
    token_algorithm text DEFAULT 'random'::text NOT NULL,
    max_attempts_per_token bigint DEFAULT 3 NOT NULL,
    max_requests_per_hour bigint DEFAULT 5 NOT NULL,
    max_requests_per_ip bigint DEFAULT 10 NOT NULL,
    cooldown_minutes bigint DEFAULT 30 NOT NULL,
    require_same_ip boolean DEFAULT false,
    allow_vp_ns boolean DEFAULT true,
    block_known_proxies boolean DEFAULT false,
    cleanup_after_days bigint DEFAULT 7 NOT NULL,
    keep_audit_days bigint DEFAULT 90 NOT NULL,
    auto_cleanup_enabled boolean DEFAULT true,
    send_confirmation_email boolean DEFAULT true,
    notify_on_suspicious boolean DEFAULT true,
    log_all_attempts boolean DEFAULT true,
    email_reset_enabled boolean DEFAULT true,
    sms_reset_enabled boolean DEFAULT false,
    admin_reset_enabled boolean DEFAULT true,
    config_name text NOT NULL,
    description text NOT NULL,
    is_active boolean DEFAULT true,
    created_by bigint NOT NULL,
    last_modified_by bigint
);


ALTER TABLE public.reset_configurations OWNER TO postgres;

--
-- Name: reset_configurations_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.reset_configurations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.reset_configurations_id_seq OWNER TO postgres;

--
-- Name: reset_configurations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.reset_configurations_id_seq OWNED BY public.reset_configurations.id;


--
-- Name: security_metrics; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.security_metrics (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    event_type text NOT NULL,
    severity text NOT NULL,
    "timestamp" timestamp with time zone NOT NULL,
    date date,
    hour bigint,
    ip_address text NOT NULL,
    user_agent text NOT NULL,
    country text,
    account_id bigint,
    user_id bigint,
    description text NOT NULL,
    raw_data text,
    risk_score bigint DEFAULT 0,
    is_blocked boolean DEFAULT false,
    action_taken text,
    is_resolved boolean DEFAULT false,
    resolved_at timestamp with time zone,
    resolved_by bigint,
    resolution text,
    CONSTRAINT chk_security_metrics_hour CHECK (((hour >= 0) AND (hour <= 23))),
    CONSTRAINT chk_security_metrics_risk_score CHECK (((risk_score >= 0) AND (risk_score <= 100)))
);


ALTER TABLE public.security_metrics OWNER TO postgres;

--
-- Name: security_metrics_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.security_metrics_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.security_metrics_id_seq OWNER TO postgres;

--
-- Name: security_metrics_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.security_metrics_id_seq OWNED BY public.security_metrics.id;


--
-- Name: settlement_items; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.settlement_items (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    settlement_record_id bigint NOT NULL,
    organizer_id bigint NOT NULL,
    event_id bigint NOT NULL,
    event_status text,
    event_end_date timestamp with time zone NOT NULL,
    event_verified_at timestamp with time zone,
    has_disputes boolean DEFAULT false,
    refund_amount_issued bigint DEFAULT 0,
    chargeback_amount bigint DEFAULT 0,
    risk_hold_applied boolean DEFAULT false,
    risk_hold_reason text,
    gross_amount bigint NOT NULL,
    platform_fee_amount bigint DEFAULT 0,
    refund_deduction bigint DEFAULT 0,
    adjustment_amount bigint DEFAULT 0,
    net_amount bigint NOT NULL,
    currency text DEFAULT 'KSH'::text NOT NULL,
    status text NOT NULL,
    external_transaction_id text,
    external_reference text,
    bank_account_number text NOT NULL,
    bank_name text NOT NULL,
    bank_code text,
    account_holder_name text NOT NULL,
    processed_at timestamp with time zone,
    completed_at timestamp with time zone,
    failed_at timestamp with time zone,
    failure_reason text,
    description text NOT NULL,
    notes text
);


ALTER TABLE public.settlement_items OWNER TO postgres;

--
-- Name: settlement_items_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.settlement_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.settlement_items_id_seq OWNER TO postgres;

--
-- Name: settlement_items_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.settlement_items_id_seq OWNED BY public.settlement_items.id;


--
-- Name: settlement_payment_records; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.settlement_payment_records (
    settlement_item_id bigint NOT NULL,
    payment_record_id bigint NOT NULL
);


ALTER TABLE public.settlement_payment_records OWNER TO postgres;

--
-- Name: settlement_records; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.settlement_records (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    settlement_batch_id text NOT NULL,
    description text NOT NULL,
    status text NOT NULL,
    frequency text NOT NULL,
    trigger text NOT NULL,
    event_id bigint,
    event_completed_at timestamp with time zone,
    event_completion_verified boolean DEFAULT false,
    holding_period_days bigint DEFAULT 7 NOT NULL,
    holding_period_start_date timestamp with time zone,
    holding_period_end_date timestamp with time zone,
    earliest_payout_date timestamp with time zone,
    has_active_disputes boolean DEFAULT false,
    dispute_count bigint DEFAULT 0,
    chargeback_count bigint DEFAULT 0,
    refund_amount bigint DEFAULT 0,
    withholding_reason text,
    period_start_date timestamp with time zone NOT NULL,
    period_end_date timestamp with time zone NOT NULL,
    total_organizers bigint DEFAULT 0 NOT NULL,
    total_amount bigint DEFAULT 0 NOT NULL,
    total_payment_records bigint DEFAULT 0 NOT NULL,
    currency text DEFAULT 'KSH'::text NOT NULL,
    initiated_by bigint,
    approved_by bigint,
    approved_at timestamp with time zone,
    processed_at timestamp with time zone,
    completed_at timestamp with time zone,
    failed_at timestamp with time zone,
    external_batch_id text,
    payment_gateway_id bigint,
    notes text,
    internal_reference text,
    risk_score bigint
);


ALTER TABLE public.settlement_records OWNER TO postgres;

--
-- Name: settlement_records_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.settlement_records_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.settlement_records_id_seq OWNER TO postgres;

--
-- Name: settlement_records_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.settlement_records_id_seq OWNED BY public.settlement_records.id;


--
-- Name: support_ticket_comments; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.support_ticket_comments (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    ticket_id bigint NOT NULL,
    user_id bigint,
    comment text NOT NULL,
    is_internal boolean DEFAULT false,
    author_name character varying(255),
    author_email character varying(255)
);


ALTER TABLE public.support_ticket_comments OWNER TO postgres;

--
-- Name: support_ticket_comments_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.support_ticket_comments_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.support_ticket_comments_id_seq OWNER TO postgres;

--
-- Name: support_ticket_comments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.support_ticket_comments_id_seq OWNED BY public.support_ticket_comments.id;


--
-- Name: support_tickets; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.support_tickets (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    ticket_number character varying(20) NOT NULL,
    subject character varying(255) NOT NULL,
    description text NOT NULL,
    category character varying(50) NOT NULL,
    priority character varying(20) DEFAULT 'medium'::character varying,
    status character varying(20) DEFAULT 'open'::character varying,
    user_id bigint,
    email character varying(255) NOT NULL,
    name character varying(255) NOT NULL,
    phone_number character varying(50),
    order_id bigint,
    event_id bigint,
    organizer_id bigint,
    assigned_to_id bigint,
    resolved_at timestamp with time zone,
    resolved_by_id bigint,
    resolution_notes text,
    ai_classified boolean DEFAULT false,
    a_ipriority character varying(20),
    ai_confidence_score numeric(5,4),
    ai_reasoning text
);


ALTER TABLE public.support_tickets OWNER TO postgres;

--
-- Name: support_tickets_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.support_tickets_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.support_tickets_id_seq OWNER TO postgres;

--
-- Name: support_tickets_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.support_tickets_id_seq OWNED BY public.support_tickets.id;


--
-- Name: system_metrics; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.system_metrics (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    metric_name text NOT NULL,
    metric_type text NOT NULL,
    granularity text NOT NULL,
    "timestamp" timestamp with time zone NOT NULL,
    date date,
    hour bigint,
    day_of_week bigint,
    week bigint,
    month bigint,
    year bigint NOT NULL,
    value numeric NOT NULL,
    count bigint DEFAULT 0,
    sum numeric DEFAULT 0,
    min numeric,
    max numeric,
    event_id bigint,
    organizer_id bigint,
    account_id bigint,
    country text,
    region text,
    city text,
    dimensions text,
    tags text,
    source text DEFAULT 'system'::text NOT NULL,
    version bigint DEFAULT 1,
    is_estimate boolean DEFAULT false,
    CONSTRAINT chk_system_metrics_day_of_week CHECK (((day_of_week >= 0) AND (day_of_week <= 6))),
    CONSTRAINT chk_system_metrics_hour CHECK (((hour >= 0) AND (hour <= 23))),
    CONSTRAINT chk_system_metrics_month CHECK (((month >= 1) AND (month <= 12))),
    CONSTRAINT chk_system_metrics_week CHECK (((week >= 1) AND (week <= 53)))
);


ALTER TABLE public.system_metrics OWNER TO postgres;

--
-- Name: system_metrics_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.system_metrics_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.system_metrics_id_seq OWNER TO postgres;

--
-- Name: system_metrics_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.system_metrics_id_seq OWNED BY public.system_metrics.id;


--
-- Name: ticket_classes; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.ticket_classes (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    event_id bigint NOT NULL,
    name text NOT NULL,
    description text,
    price bigint NOT NULL,
    currency text DEFAULT 'KSH'::text NOT NULL,
    max_per_order bigint,
    min_per_order bigint DEFAULT 1,
    quantity_available bigint,
    quantity_sold bigint DEFAULT 0,
    version bigint DEFAULT 0,
    start_sale_date timestamp with time zone,
    end_sale_date timestamp with time zone,
    sales_volume bigint DEFAULT 0,
    organizer_fees_volume bigint DEFAULT 0,
    is_paused boolean DEFAULT false,
    is_hidden boolean DEFAULT false,
    sort_order bigint DEFAULT 0,
    requires_approval boolean DEFAULT false
);


ALTER TABLE public.ticket_classes OWNER TO postgres;

--
-- Name: ticket_classes_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.ticket_classes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.ticket_classes_id_seq OWNER TO postgres;

--
-- Name: ticket_classes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.ticket_classes_id_seq OWNED BY public.ticket_classes.id;


--
-- Name: ticket_orders; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.ticket_orders (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    order_id bigint NOT NULL,
    ticket_id bigint NOT NULL
);


ALTER TABLE public.ticket_orders OWNER TO postgres;

--
-- Name: ticket_orders_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.ticket_orders_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.ticket_orders_id_seq OWNER TO postgres;

--
-- Name: ticket_orders_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.ticket_orders_id_seq OWNED BY public.ticket_orders.id;


--
-- Name: ticket_transfer_histories; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.ticket_transfer_histories (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    ticket_id bigint NOT NULL,
    from_holder_name text NOT NULL,
    from_holder_email text NOT NULL,
    to_holder_name text NOT NULL,
    to_holder_email text NOT NULL,
    transferred_by bigint NOT NULL,
    transferred_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    transfer_reason text,
    ip_address text,
    user_agent text
);


ALTER TABLE public.ticket_transfer_histories OWNER TO postgres;

--
-- Name: ticket_transfer_histories_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.ticket_transfer_histories_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.ticket_transfer_histories_id_seq OWNER TO postgres;

--
-- Name: ticket_transfer_histories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.ticket_transfer_histories_id_seq OWNED BY public.ticket_transfer_histories.id;


--
-- Name: tickets; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.tickets (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    order_item_id bigint NOT NULL,
    ticket_number text NOT NULL,
    qr_code text,
    barcode_data text,
    holder_name text,
    holder_email text,
    status text DEFAULT 'active'::text,
    checked_in_at timestamp with time zone,
    checked_in_by bigint,
    used_at timestamp with time zone,
    refunded_at timestamp with time zone,
    pdf_path text
);


ALTER TABLE public.tickets OWNER TO postgres;

--
-- Name: tickets_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.tickets_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.tickets_id_seq OWNER TO postgres;

--
-- Name: tickets_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.tickets_id_seq OWNED BY public.tickets.id;


--
-- Name: timezones; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.timezones (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    display_name text NOT NULL,
    "offset" text NOT NULL,
    iana_name text NOT NULL,
    is_active boolean DEFAULT true
);


ALTER TABLE public.timezones OWNER TO postgres;

--
-- Name: timezones_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.timezones_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.timezones_id_seq OWNER TO postgres;

--
-- Name: timezones_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.timezones_id_seq OWNED BY public.timezones.id;


--
-- Name: two_factor_attempts; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.two_factor_attempts (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id bigint NOT NULL,
    success boolean NOT NULL,
    ip_address character varying(45) NOT NULL,
    user_agent text,
    failure_type character varying(50),
    attempted_at timestamp with time zone NOT NULL
);


ALTER TABLE public.two_factor_attempts OWNER TO postgres;

--
-- Name: two_factor_attempts_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.two_factor_attempts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.two_factor_attempts_id_seq OWNER TO postgres;

--
-- Name: two_factor_attempts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.two_factor_attempts_id_seq OWNED BY public.two_factor_attempts.id;


--
-- Name: two_factor_auths; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.two_factor_auths (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id bigint NOT NULL,
    enabled boolean DEFAULT false,
    secret character varying(255) NOT NULL,
    backup_codes_hash text,
    verified_at timestamp with time zone,
    last_used_at timestamp with time zone,
    method character varying(20) DEFAULT 'totp'::character varying
);


ALTER TABLE public.two_factor_auths OWNER TO postgres;

--
-- Name: two_factor_auths_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.two_factor_auths_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.two_factor_auths_id_seq OWNER TO postgres;

--
-- Name: two_factor_auths_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.two_factor_auths_id_seq OWNED BY public.two_factor_auths.id;


--
-- Name: two_factor_sessions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.two_factor_sessions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id bigint NOT NULL,
    secret character varying(255) NOT NULL,
    verified boolean DEFAULT false,
    expires_at timestamp with time zone NOT NULL,
    ip_address character varying(45),
    user_agent text
);


ALTER TABLE public.two_factor_sessions OWNER TO postgres;

--
-- Name: two_factor_sessions_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.two_factor_sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.two_factor_sessions_id_seq OWNER TO postgres;

--
-- Name: two_factor_sessions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.two_factor_sessions_id_seq OWNED BY public.two_factor_sessions.id;


--
-- Name: user_engagement_metrics; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_engagement_metrics (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    account_id bigint NOT NULL,
    date date NOT NULL,
    session_start timestamp with time zone NOT NULL,
    session_end timestamp with time zone,
    session_duration bigint DEFAULT 0,
    page_views bigint DEFAULT 0,
    events_viewed bigint DEFAULT 0,
    search_queries bigint DEFAULT 0,
    tickets_purchased bigint DEFAULT 0,
    events_bookmarked bigint DEFAULT 0,
    social_shares bigint DEFAULT 0,
    email_signups bigint DEFAULT 0,
    revenue_generated bigint DEFAULT 0,
    user_agent text NOT NULL,
    ip_address text NOT NULL,
    country text,
    city text,
    referrer_source text,
    campaign_id text,
    utm_source text,
    utm_campaign text,
    utm_medium text
);


ALTER TABLE public.user_engagement_metrics OWNER TO postgres;

--
-- Name: user_engagement_metrics_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.user_engagement_metrics_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.user_engagement_metrics_id_seq OWNER TO postgres;

--
-- Name: user_engagement_metrics_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.user_engagement_metrics_id_seq OWNED BY public.user_engagement_metrics.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    account_id bigint NOT NULL,
    first_name text NOT NULL,
    last_name text NOT NULL,
    username text NOT NULL,
    phone text NOT NULL,
    email text NOT NULL,
    password text NOT NULL,
    confirmation_code text,
    isconfirmed boolean DEFAULT false,
    role character varying(20) DEFAULT 'customer'::character varying NOT NULL,
    is_active boolean DEFAULT true,
    profile_picture text,
    email_verified boolean DEFAULT false,
    email_verified_at timestamp with time zone,
    verification_token_exp timestamp with time zone,
    refresh_token text,
    refresh_token_exp bigint,
    last_login_at bigint,
    token_version bigint DEFAULT 1
);


ALTER TABLE public.users OWNER TO postgres;

--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.users_id_seq OWNER TO postgres;

--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: venues; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.venues (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    venue_name text NOT NULL,
    venue_capacity bigint NOT NULL,
    venue_section text,
    venue_type text NOT NULL,
    venue_location text NOT NULL,
    address text,
    city text,
    state text,
    country text,
    zip_code text,
    parking_available boolean DEFAULT true,
    parking_capacity bigint,
    is_accessible boolean DEFAULT true,
    has_wifi boolean DEFAULT false,
    has_catering boolean DEFAULT false,
    contact_email text,
    contact_phone text,
    website text
);


ALTER TABLE public.venues OWNER TO postgres;

--
-- Name: venues_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.venues_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.venues_id_seq OWNER TO postgres;

--
-- Name: venues_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.venues_id_seq OWNED BY public.venues.id;


--
-- Name: waitlist_entries; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.waitlist_entries (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    event_id bigint NOT NULL,
    ticket_class_id bigint,
    email text NOT NULL,
    name text NOT NULL,
    phone text,
    quantity bigint DEFAULT 1 NOT NULL,
    status text DEFAULT 'waiting'::text,
    notified_at timestamp with time zone,
    converted_at timestamp with time zone,
    expires_at timestamp with time zone,
    priority bigint DEFAULT 0,
    session_id text,
    user_id bigint
);


ALTER TABLE public.waitlist_entries OWNER TO postgres;

--
-- Name: waitlist_entries_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.waitlist_entries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.waitlist_entries_id_seq OWNER TO postgres;

--
-- Name: waitlist_entries_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.waitlist_entries_id_seq OWNED BY public.waitlist_entries.id;


--
-- Name: webhook_logs; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.webhook_logs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    provider text NOT NULL,
    event_id text NOT NULL,
    event_type text NOT NULL,
    status text DEFAULT 'received'::text NOT NULL,
    payload text NOT NULL,
    headers text,
    request_method text DEFAULT 'POST'::text,
    request_path text,
    processed_at timestamp with time zone,
    processing_time bigint,
    retry_count bigint DEFAULT 0,
    last_retry_at timestamp with time zone,
    success boolean DEFAULT false,
    error_message text,
    stack_trace text,
    order_id bigint,
    payment_transaction_id bigint,
    payment_record_id bigint,
    account_id bigint,
    organizer_id bigint,
    external_transaction_id text,
    external_reference text,
    signature_valid boolean DEFAULT false,
    signature_header text,
    ip_address text,
    user_agent text,
    idempotency_key text,
    is_duplicate boolean DEFAULT false,
    environment text DEFAULT 'production'::text,
    api_version text,
    notes text,
    response_status bigint DEFAULT 200,
    response_body text
);


ALTER TABLE public.webhook_logs OWNER TO postgres;

--
-- Name: webhook_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.webhook_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.webhook_logs_id_seq OWNER TO postgres;

--
-- Name: webhook_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.webhook_logs_id_seq OWNED BY public.webhook_logs.id;


--
-- Name: account_activities id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.account_activities ALTER COLUMN id SET DEFAULT nextval('public.account_activities_id_seq'::regclass);


--
-- Name: account_payment_gateways id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.account_payment_gateways ALTER COLUMN id SET DEFAULT nextval('public.account_payment_gateways_id_seq'::regclass);


--
-- Name: accounts id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.accounts ALTER COLUMN id SET DEFAULT nextval('public.accounts_id_seq'::regclass);


--
-- Name: attendees id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.attendees ALTER COLUMN id SET DEFAULT nextval('public.attendees_id_seq'::regclass);


--
-- Name: currencies id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.currencies ALTER COLUMN id SET DEFAULT nextval('public.currencies_id_seq'::regclass);


--
-- Name: date_formats id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.date_formats ALTER COLUMN id SET DEFAULT nextval('public.date_formats_id_seq'::regclass);


--
-- Name: date_time_formats id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.date_time_formats ALTER COLUMN id SET DEFAULT nextval('public.date_time_formats_id_seq'::regclass);


--
-- Name: email_verifications id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.email_verifications ALTER COLUMN id SET DEFAULT nextval('public.email_verifications_id_seq'::regclass);


--
-- Name: event_images id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.event_images ALTER COLUMN id SET DEFAULT nextval('public.event_images_id_seq'::regclass);


--
-- Name: event_metrics id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.event_metrics ALTER COLUMN id SET DEFAULT nextval('public.event_metrics_id_seq'::regclass);


--
-- Name: event_stats id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.event_stats ALTER COLUMN id SET DEFAULT nextval('public.event_stats_id_seq'::regclass);


--
-- Name: event_venues id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.event_venues ALTER COLUMN id SET DEFAULT nextval('public.event_venues_id_seq'::regclass);


--
-- Name: events id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.events ALTER COLUMN id SET DEFAULT nextval('public.events_id_seq'::regclass);


--
-- Name: login_history id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.login_history ALTER COLUMN id SET DEFAULT nextval('public.login_history_id_seq'::regclass);


--
-- Name: notification_preferences id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notification_preferences ALTER COLUMN id SET DEFAULT nextval('public.notification_preferences_id_seq'::regclass);


--
-- Name: order_items id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_items ALTER COLUMN id SET DEFAULT nextval('public.order_items_id_seq'::regclass);


--
-- Name: orders id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.orders ALTER COLUMN id SET DEFAULT nextval('public.orders_id_seq'::regclass);


--
-- Name: organizers id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.organizers ALTER COLUMN id SET DEFAULT nextval('public.organizers_id_seq'::regclass);


--
-- Name: password_reset_attempts id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.password_reset_attempts ALTER COLUMN id SET DEFAULT nextval('public.password_reset_attempts_id_seq'::regclass);


--
-- Name: password_resets id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.password_resets ALTER COLUMN id SET DEFAULT nextval('public.password_resets_id_seq'::regclass);


--
-- Name: payment_gateways id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payment_gateways ALTER COLUMN id SET DEFAULT nextval('public.payment_gateways_id_seq'::regclass);


--
-- Name: payment_methods id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payment_methods ALTER COLUMN id SET DEFAULT nextval('public.payment_methods_id_seq'::regclass);


--
-- Name: payment_records id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payment_records ALTER COLUMN id SET DEFAULT nextval('public.payment_records_id_seq'::regclass);


--
-- Name: payment_transactions id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payment_transactions ALTER COLUMN id SET DEFAULT nextval('public.payment_transactions_id_seq'::regclass);


--
-- Name: payout_accounts id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payout_accounts ALTER COLUMN id SET DEFAULT nextval('public.payout_accounts_id_seq'::regclass);


--
-- Name: promotion_rules id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.promotion_rules ALTER COLUMN id SET DEFAULT nextval('public.promotion_rules_id_seq'::regclass);


--
-- Name: promotion_usages id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.promotion_usages ALTER COLUMN id SET DEFAULT nextval('public.promotion_usages_id_seq'::regclass);


--
-- Name: promotions id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.promotions ALTER COLUMN id SET DEFAULT nextval('public.promotions_id_seq'::regclass);


--
-- Name: recovery_codes id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.recovery_codes ALTER COLUMN id SET DEFAULT nextval('public.recovery_codes_id_seq'::regclass);


--
-- Name: refund_line_items id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.refund_line_items ALTER COLUMN id SET DEFAULT nextval('public.refund_line_items_id_seq'::regclass);


--
-- Name: refund_records id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.refund_records ALTER COLUMN id SET DEFAULT nextval('public.refund_records_id_seq'::regclass);


--
-- Name: reserved_tickets id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.reserved_tickets ALTER COLUMN id SET DEFAULT nextval('public.reserved_tickets_id_seq'::regclass);


--
-- Name: reset_configurations id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.reset_configurations ALTER COLUMN id SET DEFAULT nextval('public.reset_configurations_id_seq'::regclass);


--
-- Name: security_metrics id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.security_metrics ALTER COLUMN id SET DEFAULT nextval('public.security_metrics_id_seq'::regclass);


--
-- Name: settlement_items id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.settlement_items ALTER COLUMN id SET DEFAULT nextval('public.settlement_items_id_seq'::regclass);


--
-- Name: settlement_records id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.settlement_records ALTER COLUMN id SET DEFAULT nextval('public.settlement_records_id_seq'::regclass);


--
-- Name: support_ticket_comments id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.support_ticket_comments ALTER COLUMN id SET DEFAULT nextval('public.support_ticket_comments_id_seq'::regclass);


--
-- Name: support_tickets id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.support_tickets ALTER COLUMN id SET DEFAULT nextval('public.support_tickets_id_seq'::regclass);


--
-- Name: system_metrics id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.system_metrics ALTER COLUMN id SET DEFAULT nextval('public.system_metrics_id_seq'::regclass);


--
-- Name: ticket_classes id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ticket_classes ALTER COLUMN id SET DEFAULT nextval('public.ticket_classes_id_seq'::regclass);


--
-- Name: ticket_orders id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ticket_orders ALTER COLUMN id SET DEFAULT nextval('public.ticket_orders_id_seq'::regclass);


--
-- Name: ticket_transfer_histories id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ticket_transfer_histories ALTER COLUMN id SET DEFAULT nextval('public.ticket_transfer_histories_id_seq'::regclass);


--
-- Name: tickets id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tickets ALTER COLUMN id SET DEFAULT nextval('public.tickets_id_seq'::regclass);


--
-- Name: timezones id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.timezones ALTER COLUMN id SET DEFAULT nextval('public.timezones_id_seq'::regclass);


--
-- Name: two_factor_attempts id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.two_factor_attempts ALTER COLUMN id SET DEFAULT nextval('public.two_factor_attempts_id_seq'::regclass);


--
-- Name: two_factor_auths id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.two_factor_auths ALTER COLUMN id SET DEFAULT nextval('public.two_factor_auths_id_seq'::regclass);


--
-- Name: two_factor_sessions id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.two_factor_sessions ALTER COLUMN id SET DEFAULT nextval('public.two_factor_sessions_id_seq'::regclass);


--
-- Name: user_engagement_metrics id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_engagement_metrics ALTER COLUMN id SET DEFAULT nextval('public.user_engagement_metrics_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: venues id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.venues ALTER COLUMN id SET DEFAULT nextval('public.venues_id_seq'::regclass);


--
-- Name: waitlist_entries id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.waitlist_entries ALTER COLUMN id SET DEFAULT nextval('public.waitlist_entries_id_seq'::regclass);


--
-- Name: webhook_logs id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.webhook_logs ALTER COLUMN id SET DEFAULT nextval('public.webhook_logs_id_seq'::regclass);


--
-- Data for Name: account_activities; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.account_activities (id, created_at, updated_at, deleted_at, account_id, user_id, action, category, description, ip_address, user_agent, success, metadata, severity, resource, resource_id, "timestamp") FROM stdin;
1	2025-12-03 00:27:40.890848+03	2025-12-03 00:27:40.890848+03	\N	1	1	account_created	auth	User account registered	[::1]:42590	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-03 00:27:40.890185+03
2	2025-12-03 03:58:50.415909+03	2025-12-03 03:58:50.415909+03	\N	1	1	login	auth	User logged in successfully	[::1]:46836	curl/8.5.0	t	\N	info		\N	2025-12-03 03:58:50.393633+03
3	2025-12-03 04:12:15.726508+03	2025-12-03 04:12:15.726508+03	\N	1	1	login	auth	User logged in successfully	[::1]:59284	curl/8.5.0	t	\N	info		\N	2025-12-03 04:12:15.711722+03
4	2025-12-03 04:16:22.997099+03	2025-12-03 04:16:22.997099+03	\N	1	1	login	auth	User logged in successfully	[::1]:42012	curl/8.5.0	t	\N	info		\N	2025-12-03 04:16:22.982105+03
5	2025-12-03 04:17:12.854783+03	2025-12-03 04:17:12.854783+03	\N	1	1	login	auth	User logged in successfully	[::1]:38064	curl/8.5.0	t	\N	info		\N	2025-12-03 04:17:12.840258+03
6	2025-12-03 04:21:01.204846+03	2025-12-03 04:21:01.204846+03	\N	1	1	login	auth	User logged in successfully	[::1]:59652	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-03 04:21:01.150701+03
7	2025-12-08 03:13:07.526793+03	2025-12-08 03:13:07.526793+03	\N	1	1	login	auth	User logged in successfully	[::1]:52608	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-08 03:13:07.489245+03
8	2025-12-08 03:13:24.936267+03	2025-12-08 03:13:24.936267+03	\N	1	1	logout	auth	User logged out	[::1]:52608	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-08 03:13:24.887349+03
9	2025-12-08 04:39:00.173136+03	2025-12-08 04:39:00.173136+03	\N	2	2	account_created	auth	User account registered	[::1]:38936	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-08 04:39:00.14649+03
10	2025-12-08 04:40:58.646134+03	2025-12-08 04:40:58.646134+03	\N	2	2	login	auth	User logged in successfully	[::1]:40272	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-08 04:40:58.62909+03
11	2025-12-08 04:41:25.747999+03	2025-12-08 04:41:25.747999+03	\N	2	2	logout	auth	User logged out	[::1]:40272	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-08 04:41:25.704111+03
12	2025-12-08 04:41:32.378042+03	2025-12-08 04:41:32.378042+03	\N	2	2	login	auth	User logged in successfully	[::1]:40272	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-08 04:41:32.364197+03
13	2025-12-08 05:56:57.982852+03	2025-12-08 05:56:57.982852+03	\N	2	2	password_reset_request	security	Password reset requested	[::1]:37172	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-08 05:56:57.967602+03
14	2025-12-08 06:15:37.354399+03	2025-12-08 06:15:37.354399+03	\N	2	2	password_reset_request	security	Password reset requested	[::1]:44702	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-08 06:15:37.334071+03
15	2025-12-08 06:16:09.942185+03	2025-12-08 06:16:09.942185+03	\N	2	2	password_reset	security	Password reset completed	[::1]:44702	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-08 06:16:09.921063+03
16	2025-12-08 06:16:34.00969+03	2025-12-08 06:16:34.00969+03	\N	2	\N	login_failed	auth	Failed login attempt: invalid password	[::1]:44702	PostmanRuntime/7.49.1	t	\N	warning		\N	2025-12-08 06:16:33.986028+03
17	2025-12-08 06:16:43.358566+03	2025-12-08 06:16:43.358566+03	\N	2	2	login	auth	User logged in successfully	[::1]:44702	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-08 06:16:43.340152+03
18	2025-12-08 06:16:58.91462+03	2025-12-08 06:16:58.91462+03	\N	2	2	logout	auth	User logged out	[::1]:44702	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-08 06:16:58.894939+03
19	2025-12-08 06:17:04.289148+03	2025-12-08 06:17:04.289148+03	\N	2	2	password_reset_request	security	Password reset requested	[::1]:44702	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-08 06:17:04.264751+03
20	2025-12-08 06:17:37.323125+03	2025-12-08 06:17:37.323125+03	\N	2	2	password_reset	security	Password reset completed	[::1]:44702	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-08 06:17:37.305161+03
21	2025-12-08 06:17:52.52858+03	2025-12-08 06:17:52.52858+03	\N	2	\N	login_failed	auth	Failed login attempt: invalid password	[::1]:44702	PostmanRuntime/7.49.1	t	\N	warning		\N	2025-12-08 06:17:52.515307+03
22	2025-12-08 06:17:56.03858+03	2025-12-08 06:17:56.03858+03	\N	2	2	login	auth	User logged in successfully	[::1]:44702	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-08 06:17:56.016404+03
23	2025-12-08 06:25:54.394937+03	2025-12-08 06:25:54.394937+03	\N	2	2	email_verified	security	Email address verified	[::1]:48432	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-08 06:25:54.377719+03
24	2025-12-08 06:44:59.643524+03	2025-12-08 06:44:59.643524+03	\N	2	2	login	auth	User logged in successfully	[::1]:52274	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-08 06:44:59.603643+03
25	2025-12-09 01:49:58.830607+03	2025-12-09 01:49:58.830607+03	\N	2	2	login	auth	User logged in successfully	[::1]:48172	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-09 01:49:58.80074+03
26	2025-12-09 02:11:10.882888+03	2025-12-09 02:11:10.882888+03	\N	2	2	2fa_enabled	security	Two-factor authentication enabled	[::1]:36996	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-09 02:11:10.86025+03
27	2025-12-09 02:18:17.995658+03	2025-12-09 02:18:17.995658+03	\N	2	2	2fa_verified	security	Two-factor authentication verified	[::1]:59650	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-09 02:18:17.964236+03
28	2025-12-09 02:23:14.18355+03	2025-12-09 02:23:14.18355+03	\N	2	2	2fa_disabled	security	Two-factor authentication disabled	[::1]:55308	PostmanRuntime/7.49.1	t	\N	warning		\N	2025-12-09 02:23:14.157555+03
29	2025-12-09 02:27:18.513567+03	2025-12-09 02:27:18.513567+03	\N	2	2	login	auth	User logged in successfully	[::1]:55558	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-09 02:27:18.457355+03
30	2025-12-09 02:43:45.299075+03	2025-12-09 02:43:45.299075+03	\N	2	2	logout	auth	User logged out	[::1]:41086	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-09 02:43:45.275381+03
31	2025-12-09 02:43:52.892014+03	2025-12-09 02:43:52.892014+03	\N	2	2	login	auth	User logged in successfully	[::1]:41086	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-09 02:43:52.851786+03
32	2025-12-09 02:54:47.099309+03	2025-12-09 02:54:47.099309+03	\N	2	2	2fa_enabled	security	Two-factor authentication enabled	[::1]:40006	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-09 02:54:47.059613+03
33	2025-12-09 02:55:02.215179+03	2025-12-09 02:55:02.215179+03	\N	2	2	logout	auth	User logged out	[::1]:40006	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-09 02:55:02.190238+03
34	2025-12-09 02:55:41.966896+03	2025-12-09 02:55:41.966896+03	\N	2	2	2fa_verified	security	Two-factor authentication verified	[::1]:40006	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-09 02:55:41.943808+03
35	2025-12-09 02:56:41.929189+03	2025-12-09 02:56:41.929189+03	\N	2	2	recovery_codes_regenerated	security	Two-factor recovery codes regenerated	[::1]:40006	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-09 02:56:41.905345+03
36	2025-12-09 03:01:41.385539+03	2025-12-09 03:01:41.385539+03	\N	2	2	2fa_disabled	security	Two-factor authentication disabled	[::1]:51518	PostmanRuntime/7.49.1	t	\N	warning		\N	2025-12-09 03:01:41.355148+03
37	2025-12-09 05:55:26.938627+03	2025-12-09 05:55:26.938627+03	\N	3	3	account_created	auth	User account registered	[::1]:59792	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-09 05:55:26.888685+03
38	2025-12-09 05:57:44.762298+03	2025-12-09 05:57:44.762298+03	\N	3	3	email_verified	security	Email address verified	[::1]:49856	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-09 05:57:44.740814+03
39	2025-12-09 06:00:09.772542+03	2025-12-09 06:00:09.772542+03	\N	3	3	login	auth	User logged in successfully	[::1]:42464	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-09 06:00:09.75346+03
40	2025-12-09 07:08:22.179208+03	2025-12-09 07:08:22.179208+03	\N	3	3	login	auth	User logged in successfully	[::1]:57596	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-09 07:08:22.160603+03
41	2025-12-09 20:34:29.737762+03	2025-12-09 20:34:29.737762+03	\N	3	3	login	auth	User logged in successfully	[::1]:33654	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-09 20:34:29.709039+03
42	2025-12-09 20:56:42.282557+03	2025-12-09 20:56:42.282557+03	\N	3	3	logout	auth	User logged out	[::1]:59370	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-09 20:56:42.264373+03
43	2025-12-09 20:57:42.385203+03	2025-12-09 20:57:42.385203+03	\N	4	4	account_created	auth	User account registered	[::1]:59370	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-09 20:57:42.369843+03
44	2025-12-09 20:58:20.197549+03	2025-12-09 20:58:20.197549+03	\N	4	4	email_verified	security	Email address verified	[::1]:59370	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-09 20:58:20.146966+03
45	2025-12-09 20:58:47.435198+03	2025-12-09 20:58:47.435198+03	\N	4	4	login	auth	User logged in successfully	[::1]:59370	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-09 20:58:47.363097+03
46	2025-12-09 20:58:48.901171+03	2025-12-09 20:58:48.901171+03	\N	4	4	logout	auth	User logged out	[::1]:59370	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-09 20:58:48.800736+03
47	2025-12-09 21:06:45.25821+03	2025-12-09 21:06:45.25821+03	\N	4	4	login	auth	User logged in successfully	[::1]:51588	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-09 21:06:45.241222+03
48	2025-12-09 21:08:34.280652+03	2025-12-09 21:08:34.280652+03	\N	4	4	login	auth	User logged in successfully	[::1]:56596	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-09 21:08:34.236066+03
49	2025-12-09 21:08:39.921553+03	2025-12-09 21:08:39.921553+03	\N	4	4	logout	auth	User logged out	[::1]:56596	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-09 21:08:39.876453+03
50	2025-12-09 21:08:45.371785+03	2025-12-09 21:08:45.371785+03	\N	4	4	login	auth	User logged in successfully	[::1]:56596	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-09 21:08:45.353555+03
51	2025-12-09 21:10:15.534018+03	2025-12-09 21:10:15.534018+03	\N	4	4	login	auth	User logged in successfully	[::1]:40480	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-09 21:10:15.507617+03
52	2025-12-09 21:12:11.114759+03	2025-12-09 21:12:11.114759+03	\N	4	4	login	auth	User logged in successfully	[::1]:40934	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-09 21:12:11.098588+03
53	2025-12-09 21:23:20.747612+03	2025-12-09 21:23:20.747612+03	\N	3	3	login	auth	User logged in successfully	[::1]:56326	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-09 21:23:20.734847+03
54	2025-12-09 21:23:37.216221+03	2025-12-09 21:23:37.216221+03	\N	4	4	login	auth	User logged in successfully	[::1]:56326	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-09 21:23:37.202135+03
55	2025-12-09 23:42:19.911327+03	2025-12-09 23:42:19.911327+03	\N	4	4	login	auth	User logged in successfully	[::1]:55226	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-09 23:42:19.891111+03
56	2025-12-10 00:38:45.184506+03	2025-12-10 00:38:45.184506+03	\N	3	3	login	auth	User logged in successfully	[::1]:48936	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-10 00:38:45.167241+03
57	2025-12-10 00:51:35.430815+03	2025-12-10 00:51:35.430815+03	\N	4	4	login	auth	User logged in successfully	[::1]:49796	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-10 00:51:35.413191+03
58	2025-12-10 01:27:12.590385+03	2025-12-10 01:27:12.590385+03	\N	5	5	account_created	auth	User account registered	[::1]:47204	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-10 01:27:12.563259+03
59	2025-12-10 01:27:50.476946+03	2025-12-10 01:27:50.476946+03	\N	5	5	email_verified	security	Email address verified	[::1]:47204	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-10 01:27:50.449688+03
60	2025-12-10 01:28:13.681807+03	2025-12-10 01:28:13.681807+03	\N	5	5	login	auth	User logged in successfully	[::1]:47204	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-10 01:28:13.667792+03
61	2025-12-10 01:28:33.2303+03	2025-12-10 01:28:33.2303+03	\N	4	4	login	auth	User logged in successfully	[::1]:47204	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-10 01:28:33.191107+03
62	2025-12-10 01:44:11.354678+03	2025-12-10 01:44:11.354678+03	\N	3	3	login	auth	User logged in successfully	[::1]:40868	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-10 01:44:11.310578+03
63	2025-12-10 04:44:03.158758+03	2025-12-10 04:44:03.158758+03	\N	3	3	login	auth	User logged in successfully	[::1]:49102	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-10 04:44:03.139637+03
64	2025-12-10 05:45:38.030166+03	2025-12-10 05:45:38.030166+03	\N	3	3	login	auth	User logged in successfully	[::1]:45984	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-10 05:45:37.873831+03
65	2025-12-10 05:58:05.562526+03	2025-12-10 05:58:05.562526+03	\N	3	3	login	auth	User logged in successfully	[::1]:36080	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-10 05:58:05.546958+03
66	2025-12-10 07:02:23.409325+03	2025-12-10 07:02:23.409325+03	\N	3	3	login	auth	User logged in successfully	[::1]:38390	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-10 07:02:23.387341+03
67	2025-12-10 22:09:13.650028+03	2025-12-10 22:09:13.650028+03	\N	3	3	login	auth	User logged in successfully	[::1]:42504	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-10 22:09:13.593037+03
68	2025-12-11 00:12:50.056211+03	2025-12-11 00:12:50.056211+03	\N	3	3	login	auth	User logged in successfully	[::1]:53738	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-11 00:12:49.990264+03
69	2025-12-11 00:15:28.524489+03	2025-12-11 00:15:28.524489+03	\N	3	\N	profile_updated	profile	Profile information updated	[::1]:45580		t	\N	info		\N	2025-12-11 00:15:28.486566+03
70	2025-12-11 00:17:03.7056+03	2025-12-11 00:17:03.7056+03	\N	1	1	login	auth	User logged in successfully	[::1]:51510	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-11 00:17:03.687988+03
71	2025-12-11 00:17:51.212683+03	2025-12-11 00:17:51.212683+03	\N	1	\N	account_deleted	general	Account deactivated	[::1]:51510		t	\N	info		\N	2025-12-11 00:17:51.193352+03
72	2025-12-11 00:18:20.811368+03	2025-12-11 00:18:20.811368+03	\N	1	1	login	auth	User logged in successfully	[::1]:51510	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-11 00:18:20.795925+03
73	2025-12-11 01:55:23.476526+03	2025-12-11 01:55:23.476526+03	\N	3	3	login	auth	User logged in successfully	[::1]:49560	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-11 01:55:23.441632+03
74	2025-12-11 02:02:06.174779+03	2025-12-11 02:02:06.174779+03	\N	3	\N	address_updated	profile	Address information updated	[::1]:41492		t	\N	info		\N	2025-12-11 02:02:06.145517+03
75	2025-12-11 02:04:03.587538+03	2025-12-11 02:04:03.587538+03	\N	3	\N	address_updated	profile	Address information updated	[::1]:41492		t	\N	info		\N	2025-12-11 02:04:03.555085+03
76	2025-12-11 02:27:18.615664+03	2025-12-11 02:27:18.615664+03	\N	3	\N	preferences_updated	settings	Account preferences updated	[::1]:45008		t	\N	info		\N	2025-12-11 02:27:18.589072+03
77	2025-12-11 02:52:10.148224+03	2025-12-11 02:52:10.148224+03	\N	3	3	login	auth	User logged in successfully	[::1]:51162	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-11 02:52:10.134339+03
78	2025-12-11 03:07:35.941289+03	2025-12-11 03:07:35.941289+03	\N	3	\N	preferences_updated	settings	Account preferences updated	[::1]:51856		t	\N	info		\N	2025-12-11 03:07:35.916465+03
79	2025-12-11 04:38:12.92856+03	2025-12-11 04:38:12.92856+03	\N	3	3	login	auth	User logged in successfully	[::1]:45492	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-11 04:38:12.914093+03
80	2025-12-11 04:49:49.342968+03	2025-12-11 04:49:49.342968+03	\N	3	\N	password_changed	security	Password changed successfully	[::1]:56360		t	\N	info		\N	2025-12-11 04:49:49.327316+03
81	2025-12-11 04:49:49.374257+03	2025-12-11 04:49:49.374257+03	\N	3	\N	password_changed	security	Security event: password_changed	[::1]:56360	PostmanRuntime/7.49.1	t	\N	warning		\N	2025-12-11 04:49:49.357466+03
82	2025-12-11 04:50:00.467377+03	2025-12-11 04:50:00.467377+03	\N	3	\N	login_failed	auth	Failed login attempt: invalid password	[::1]:56360	PostmanRuntime/7.49.1	t	\N	warning		\N	2025-12-11 04:50:00.453286+03
83	2025-12-11 04:50:12.530596+03	2025-12-11 04:50:12.530596+03	\N	3	\N	password_changed	security	Password changed successfully	[::1]:56360		t	\N	info		\N	2025-12-11 04:50:12.510677+03
84	2025-12-11 04:50:12.55679+03	2025-12-11 04:50:12.55679+03	\N	3	\N	password_changed	security	Security event: password_changed	[::1]:56360	PostmanRuntime/7.49.1	t	\N	warning		\N	2025-12-11 04:50:12.542386+03
85	2025-12-11 04:50:18.449224+03	2025-12-11 04:50:18.449224+03	\N	3	3	login	auth	User logged in successfully	[::1]:56360	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-11 04:50:18.431989+03
86	2025-12-11 04:51:51.865235+03	2025-12-11 04:51:51.865235+03	\N	3	3	login	auth	User logged in successfully	[::1]:56360	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-11 04:51:51.847696+03
87	2025-12-11 04:52:00.78591+03	2025-12-11 04:52:00.78591+03	\N	4	4	login	auth	User logged in successfully	[::1]:56360	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-11 04:52:00.77151+03
88	2025-12-11 04:59:00.975045+03	2025-12-11 04:59:00.975045+03	\N	1	\N	account_unlocked	security	Security event: account_unlocked	[::1]:60092	PostmanRuntime/7.49.1	t	\N	warning		\N	2025-12-11 04:59:00.952656+03
89	2025-12-11 04:59:01.013498+03	2025-12-11 04:59:01.013498+03	\N	1	\N	account_unlocked	general	Account unlocked by admin (User ID: 4)	[::1]:60092		t	\N	info		\N	2025-12-11 04:59:00.992565+03
90	2025-12-11 04:59:16.245847+03	2025-12-11 04:59:16.245847+03	\N	3	3	login	auth	User logged in successfully	[::1]:60092	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-11 04:59:16.220662+03
91	2025-12-11 05:24:23.453008+03	2025-12-11 05:24:23.453008+03	\N	2	2	login	auth	User logged in successfully	[::1]:51290	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-11 05:24:23.429976+03
92	2025-12-11 05:26:35.5921+03	2025-12-11 05:26:35.5921+03	\N	3	3	login	auth	User logged in successfully	[::1]:41310	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-11 05:26:35.574509+03
93	2025-12-11 05:31:34.629822+03	2025-12-11 05:31:34.629822+03	\N	2	2	login	auth	User logged in successfully	[::1]:37622	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-11 05:31:34.603312+03
94	2025-12-12 01:49:24.529087+03	2025-12-12 01:49:24.529087+03	\N	3	3	login	auth	User logged in successfully	[::1]:41418	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-12 01:49:24.503599+03
95	2025-12-12 05:24:24.533711+03	2025-12-12 05:24:24.533711+03	\N	3	3	login	auth	User logged in successfully	[::1]:48578	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-12 05:24:24.48979+03
96	2025-12-12 05:28:05.059037+03	2025-12-12 05:28:05.059037+03	\N	2	2	login	auth	User logged in successfully	[::1]:56138	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-12 05:28:05.043067+03
97	2025-12-12 06:31:49.131047+03	2025-12-12 06:31:49.131047+03	\N	2	2	login	auth	User logged in successfully	[::1]:40782	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-12 06:31:49.116104+03
98	2025-12-15 06:23:33.740293+03	2025-12-15 06:23:33.740293+03	\N	2	2	login	auth	User logged in successfully	[::1]:59658	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-15 06:23:33.721488+03
99	2025-12-16 04:22:49.743746+03	2025-12-16 04:22:49.743746+03	\N	2	2	login	auth	User logged in successfully	[::1]:50112	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-16 04:22:49.723675+03
100	2025-12-16 05:09:33.823097+03	2025-12-16 05:09:33.823097+03	\N	2	2	login	auth	User logged in successfully	[::1]:48612	PostmanRuntime/7.49.1	t	\N	info		\N	2025-12-16 05:09:33.69792+03
101	2025-12-16 05:32:29.717242+03	2025-12-16 05:32:29.717242+03	\N	2	2	login	auth	User logged in successfully	[::1]:34744	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-16 05:32:29.644794+03
102	2025-12-16 06:24:32.521088+03	2025-12-16 06:24:32.521088+03	\N	2	2	login	auth	User logged in successfully	[::1]:59036	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-16 06:24:32.497951+03
103	2025-12-16 07:53:56.081353+03	2025-12-16 07:53:56.081353+03	\N	2	2	login	auth	User logged in successfully	[::1]:49694	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-16 07:53:56.059821+03
104	2025-12-16 08:06:54.380622+03	2025-12-16 08:06:54.380622+03	\N	4	4	login	auth	User logged in successfully	[::1]:50524	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-16 08:06:54.362219+03
105	2025-12-22 01:50:14.046452+03	2025-12-22 01:50:14.046452+03	\N	2	2	login	auth	User logged in successfully	[::1]:37822	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-22 01:50:14.029893+03
106	2025-12-22 01:53:02.371239+03	2025-12-22 01:53:02.371239+03	\N	4	4	login	auth	User logged in successfully	[::1]:35580	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-22 01:53:02.353852+03
107	2025-12-22 03:34:08.944721+03	2025-12-22 03:34:08.944721+03	\N	4	4	login	auth	User logged in successfully	[::1]:33074	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-22 03:34:08.912921+03
108	2025-12-22 05:24:05.65745+03	2025-12-22 05:24:05.65745+03	\N	3	3	login	auth	User logged in successfully	[::1]:55030	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-22 05:24:05.626043+03
109	2025-12-22 05:46:59.452471+03	2025-12-22 05:46:59.452471+03	\N	3	3	login	auth	User logged in successfully	[::1]:57160	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-22 05:46:59.403361+03
110	2025-12-22 06:12:48.01802+03	2025-12-22 06:12:48.01802+03	\N	2	2	login	auth	User logged in successfully	[::1]:50848	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-22 06:12:48.001596+03
111	2025-12-22 07:09:16.683169+03	2025-12-22 07:09:16.683169+03	\N	4	4	login	auth	User logged in successfully	[::1]:50940	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-22 07:09:16.660608+03
112	2025-12-22 07:11:35.155778+03	2025-12-22 07:11:35.155778+03	\N	2	2	login	auth	User logged in successfully	[::1]:50940	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-22 07:11:35.133548+03
113	2025-12-23 13:06:43.368001+03	2025-12-23 13:06:43.368001+03	\N	2	2	login	auth	User logged in successfully	[::1]:38442	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-23 13:06:43.337487+03
114	2025-12-23 14:18:12.248597+03	2025-12-23 14:18:12.248597+03	\N	2	2	login	auth	User logged in successfully	[::1]:39316	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-23 14:18:12.194149+03
115	2025-12-23 14:33:23.959586+03	2025-12-23 14:33:23.959586+03	\N	2	2	ticket_transferred	ticket	Ticket TKT-2-1-1-0-1765855030 transferred from kamausimon217@gmail.com to harshasg10@gmail.com	[::1]:34296	PostmanRuntime/7.51.0	t	\N	info	ticket	\N	2025-12-23 14:33:23.925872+03
116	2025-12-23 14:39:17.099282+03	2025-12-23 14:39:17.099282+03	\N	2	2	ticket_transferred	ticket	Ticket TKT-2-1-1-1-1765855030 transferred from kamausimon217@gmail.com to harshasg10@gmail.com	[::1]:48340	PostmanRuntime/7.51.0	t	\N	info	ticket	\N	2025-12-23 14:39:17.085421+03
117	2025-12-23 14:57:46.702551+03	2025-12-23 14:57:46.702551+03	\N	3	3	login	auth	User logged in successfully	[::1]:50148	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-23 14:57:46.681104+03
118	2025-12-24 01:26:16.906145+03	2025-12-24 01:26:16.906145+03	\N	3	3	login	auth	User logged in successfully	[::1]:54464	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-24 01:26:16.863983+03
119	2025-12-24 02:34:16.614669+03	2025-12-24 02:34:16.614669+03	\N	3	3	login	auth	User logged in successfully	[::1]:34136	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-24 02:34:16.595981+03
120	2025-12-24 03:55:12.029309+03	2025-12-24 03:55:12.029309+03	\N	3	3	login	auth	User logged in successfully	[::1]:35634	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-24 03:55:12.013995+03
121	2025-12-24 04:31:38.738547+03	2025-12-24 04:31:38.738547+03	\N	3	3	login	auth	User logged in successfully	[::1]:51086	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-24 04:31:38.715809+03
122	2025-12-24 09:10:14.908036+03	2025-12-24 09:10:14.908036+03	\N	3	3	login	auth	User logged in successfully	[::1]:56662	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-24 09:10:14.785729+03
123	2025-12-24 10:13:39.426439+03	2025-12-24 10:13:39.426439+03	\N	3	3	login	auth	User logged in successfully	[::1]:53098	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-24 10:13:39.336842+03
124	2025-12-24 10:50:10.073494+03	2025-12-24 10:50:10.073494+03	\N	4	4	login	auth	User logged in successfully	[::1]:52786	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-24 10:50:10.050305+03
125	2025-12-24 10:50:52.633643+03	2025-12-24 10:50:52.633643+03	\N	3	3	login	auth	User logged in successfully	[::1]:52786	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-24 10:50:52.616736+03
126	2025-12-24 11:14:44.334639+03	2025-12-24 11:14:44.334639+03	\N	3	3	login	auth	User logged in successfully	[::1]:58690	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-24 11:14:44.307583+03
127	2025-12-24 11:17:58.626967+03	2025-12-24 11:17:58.626967+03	\N	3	3	login	auth	User logged in successfully	[::1]:53768	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-24 11:17:58.610292+03
128	2025-12-24 11:30:46.356554+03	2025-12-24 11:30:46.356554+03	\N	3	3	login	auth	User logged in successfully	[::1]:35838	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-24 11:30:46.340273+03
129	2025-12-24 12:11:10.770655+03	2025-12-24 12:11:10.770655+03	\N	3	3	login	auth	User logged in successfully	[::1]:49494	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-24 12:11:10.708829+03
130	2025-12-28 06:34:46.430901+03	2025-12-28 06:34:46.430901+03	\N	2	2	login	auth	User logged in successfully	[::1]:46380	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-28 06:34:46.377505+03
131	2025-12-28 09:13:49.501699+03	2025-12-28 09:13:49.501699+03	\N	2	2	login	auth	User logged in successfully	[::1]:60496	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-28 09:13:49.474636+03
132	2025-12-28 09:24:34.997699+03	2025-12-28 09:24:34.997699+03	\N	2	2	login	auth	User logged in successfully	[::1]:33700	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-28 09:24:34.981525+03
133	2025-12-28 09:29:13.573098+03	2025-12-28 09:29:13.573098+03	\N	2	2	login	auth	User logged in successfully	[::1]:48924	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-28 09:29:13.553964+03
134	2025-12-28 09:46:37.033367+03	2025-12-28 09:46:37.033367+03	\N	2	2	login	auth	User logged in successfully	[::1]:50236	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-28 09:46:37.019269+03
135	2025-12-28 11:10:38.625622+03	2025-12-28 11:10:38.625622+03	\N	2	2	login	auth	User logged in successfully	[::1]:37882	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-28 11:10:38.60727+03
136	2025-12-28 18:19:56.934837+03	2025-12-28 18:19:56.934837+03	\N	2	2	login	auth	User logged in successfully	[::1]:44662	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-28 18:19:56.892608+03
137	2025-12-28 18:48:21.338089+03	2025-12-28 18:48:21.338089+03	\N	2	2	login	auth	User logged in successfully	[::1]:47434	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-28 18:48:21.312+03
138	2025-12-29 13:19:29.139933+03	2025-12-29 13:19:29.139933+03	\N	2	2	login	auth	User logged in successfully	[::1]:40062	Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36	t	\N	info		\N	2025-12-29 13:19:29.106565+03
139	2025-12-29 13:23:41.243509+03	2025-12-29 13:23:41.243509+03	\N	2	2	login	auth	User logged in successfully	[::1]:40062	Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36	t	\N	info		\N	2025-12-29 13:23:41.231402+03
140	2025-12-29 13:23:54.012823+03	2025-12-29 13:23:54.012823+03	\N	2	2	login	auth	User logged in successfully	[::1]:40062	Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36	t	\N	info		\N	2025-12-29 13:23:54.000627+03
141	2025-12-29 13:24:37.362625+03	2025-12-29 13:24:37.362625+03	\N	2	2	login	auth	User logged in successfully	[::1]:40062	Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36	t	\N	info		\N	2025-12-29 13:24:37.346865+03
142	2025-12-29 13:25:49.149733+03	2025-12-29 13:25:49.149733+03	\N	2	2	login	auth	User logged in successfully	[::1]:40062	Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36	t	\N	info		\N	2025-12-29 13:25:49.120409+03
143	2025-12-29 13:26:04.280164+03	2025-12-29 13:26:04.280164+03	\N	2	2	login	auth	User logged in successfully	[::1]:40062	Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36	t	\N	info		\N	2025-12-29 13:26:04.26294+03
144	2025-12-29 13:26:39.478076+03	2025-12-29 13:26:39.478076+03	\N	2	2	login	auth	User logged in successfully	[::1]:40062	Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36	t	\N	info		\N	2025-12-29 13:26:39.462831+03
145	2025-12-29 13:27:06.770015+03	2025-12-29 13:27:06.770015+03	\N	1	\N	login_failed	auth	Failed login attempt: invalid password	[::1]:45610	curl/8.5.0	t	\N	warning		\N	2025-12-29 13:27:06.747596+03
146	2025-12-29 13:27:29.300589+03	2025-12-29 13:27:29.300589+03	\N	6	6	account_created	auth	User account registered	[::1]:34730	curl/8.5.0	t	\N	info		\N	2025-12-29 13:27:29.268035+03
147	2025-12-29 13:27:39.650875+03	2025-12-29 13:27:39.650875+03	\N	6	6	login	auth	User logged in successfully	[::1]:43206	curl/8.5.0	t	\N	info		\N	2025-12-29 13:27:39.634464+03
148	2025-12-29 13:30:26.867592+03	2025-12-29 13:30:26.867592+03	\N	2	2	login	auth	User logged in successfully	[::1]:40062	Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36	t	\N	info		\N	2025-12-29 13:30:26.853605+03
149	2025-12-29 13:30:33.785701+03	2025-12-29 13:30:33.785701+03	\N	2	2	login	auth	User logged in successfully	[::1]:40062	Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36	t	\N	info		\N	2025-12-29 13:30:33.747108+03
150	2025-12-29 13:31:26.722906+03	2025-12-29 13:31:26.722906+03	\N	2	2	login	auth	User logged in successfully	[::1]:40062	Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36	t	\N	info		\N	2025-12-29 13:31:26.707369+03
151	2025-12-29 15:13:16.552396+03	2025-12-29 15:13:16.552396+03	\N	2	2	login	auth	User logged in successfully	[::1]:56814	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	t	\N	info		\N	2025-12-29 15:13:16.520981+03
152	2025-12-29 15:21:35.438473+03	2025-12-29 15:21:35.438473+03	\N	2	2	login	auth	User logged in successfully	[::1]:49852	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	t	\N	info		\N	2025-12-29 15:21:35.25216+03
153	2025-12-29 15:39:36.000858+03	2025-12-29 15:39:36.000858+03	\N	2	2	login	auth	User logged in successfully	[::1]:40974	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Code/1.107.1 Chrome/142.0.7444.175 Electron/39.2.3 Safari/537.36	t	\N	info		\N	2025-12-29 15:39:35.977392+03
154	2025-12-29 16:20:56.004858+03	2025-12-29 16:20:56.004858+03	\N	2	2	login	auth	User logged in successfully	[::1]:57028	PostmanRuntime/7.51.0	t	\N	info		\N	2025-12-29 16:20:55.985984+03
155	2026-01-08 12:09:03.332387+03	2026-01-08 12:09:03.332387+03	\N	2	2	login	auth	User logged in successfully	[::1]:47256	PostmanRuntime/7.51.0	t	\N	info		\N	2026-01-08 12:09:03.308334+03
156	2026-01-08 12:55:06.962142+03	2026-01-08 12:55:06.962142+03	\N	4	4	login	auth	User logged in successfully	[::1]:36732	PostmanRuntime/7.51.0	t	\N	info		\N	2026-01-08 12:55:06.931507+03
157	2026-01-08 13:44:30.02838+03	2026-01-08 13:44:30.02838+03	\N	2	2	login	auth	User logged in successfully	[::1]:44190	PostmanRuntime/7.51.0	t	\N	info		\N	2026-01-08 13:44:30.010696+03
158	2026-01-08 13:47:27.83794+03	2026-01-08 13:47:27.83794+03	\N	4	4	login	auth	User logged in successfully	[::1]:32982	PostmanRuntime/7.51.0	t	\N	info		\N	2026-01-08 13:47:27.817869+03
159	2026-01-08 19:03:32.314155+03	2026-01-08 19:03:32.314155+03	\N	4	4	login	auth	User logged in successfully	[::1]:45670	PostmanRuntime/7.51.0	t	\N	info		\N	2026-01-08 19:03:31.785394+03
160	2026-01-08 19:03:44.332819+03	2026-01-08 19:03:44.332819+03	\N	2	2	login	auth	User logged in successfully	[::1]:45670	PostmanRuntime/7.51.0	t	\N	info		\N	2026-01-08 19:03:44.30873+03
161	2026-01-08 19:10:44.203123+03	2026-01-08 19:10:44.203123+03	\N	2	2	login	auth	User logged in successfully	[::1]:51714	PostmanRuntime/7.51.0	t	\N	info		\N	2026-01-08 19:10:44.189503+03
162	2026-01-08 20:13:08.599614+03	2026-01-08 20:13:08.599614+03	\N	2	2	login	auth	User logged in successfully	[::1]:41398	PostmanRuntime/7.51.0	t	\N	info		\N	2026-01-08 20:13:08.585635+03
163	2026-01-08 21:16:46.552597+03	2026-01-08 21:16:46.552597+03	\N	2	2	login	auth	User logged in successfully	[::1]:58828	PostmanRuntime/7.51.0	t	\N	info		\N	2026-01-08 21:16:46.533034+03
164	2026-01-09 09:31:22.564627+03	2026-01-09 09:31:22.564627+03	\N	4	4	login	auth	User logged in successfully	[::1]:60336	PostmanRuntime/7.51.0	t	\N	info		\N	2026-01-09 09:31:22.539003+03
165	2026-01-09 09:51:46.697067+03	2026-01-09 09:51:46.697067+03	\N	4	4	login	auth	User logged in successfully	[::1]:33748	PostmanRuntime/7.51.0	t	\N	info		\N	2026-01-09 09:51:46.66013+03
166	2026-01-09 10:58:36.818297+03	2026-01-09 10:58:36.818297+03	\N	4	4	login	auth	User logged in successfully	[::1]:34030	PostmanRuntime/7.51.0	t	\N	info		\N	2026-01-09 10:58:36.799535+03
167	2026-01-09 12:40:31.333664+03	2026-01-09 12:40:31.333664+03	\N	4	4	login	auth	User logged in successfully	[::1]:60414	PostmanRuntime/7.51.0	t	\N	info		\N	2026-01-09 12:40:31.31069+03
168	2026-01-10 18:13:59.086897+03	2026-01-10 18:13:59.086897+03	\N	2	2	login	auth	User logged in successfully	[::1]:55768	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	t	\N	info		\N	2026-01-10 18:13:59.066675+03
169	2026-01-10 18:26:22.303354+03	2026-01-10 18:26:22.303354+03	\N	2	2	login	auth	User logged in successfully	172.22.16.1:51491	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	t	\N	info		\N	2026-01-10 18:26:22.265516+03
170	2026-01-10 21:24:30.85841+03	2026-01-10 21:24:30.85841+03	\N	3	3	login	auth	User logged in successfully	[::1]:38674	PostmanRuntime/7.51.0	t	\N	info		\N	2026-01-10 21:24:30.840768+03
171	2026-01-10 22:25:40.535979+03	2026-01-10 22:25:40.535979+03	\N	2	2	login	auth	User logged in successfully	172.22.16.1:58337	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	t	\N	info		\N	2026-01-10 22:25:40.521446+03
172	2026-01-10 22:29:03.915057+03	2026-01-10 22:29:03.915057+03	\N	3	3	login	auth	User logged in successfully	[::1]:52238	PostmanRuntime/7.51.0	t	\N	info		\N	2026-01-10 22:29:03.897522+03
173	2026-01-10 22:41:09.204057+03	2026-01-10 22:41:09.204057+03	\N	2	2	login	auth	User logged in successfully	172.22.16.1:58880	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	t	\N	info		\N	2026-01-10 22:41:09.189386+03
174	2026-01-10 22:48:15.858099+03	2026-01-10 22:48:15.858099+03	\N	2	2	login	auth	User logged in successfully	172.22.16.1:58880	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	t	\N	info		\N	2026-01-10 22:48:15.829625+03
175	2026-01-10 23:01:02.266357+03	2026-01-10 23:01:02.266357+03	\N	2	2	login	auth	User logged in successfully	172.22.16.1:59351	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	t	\N	info		\N	2026-01-10 23:01:02.24855+03
176	2026-01-12 12:39:08.835728+03	2026-01-12 12:39:08.835728+03	\N	2	2	login	auth	User logged in successfully	172.22.16.1:64336	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	t	\N	info		\N	2026-01-12 12:39:08.818434+03
177	2026-01-12 12:52:11.744521+03	2026-01-12 12:52:11.744521+03	\N	2	2	login	auth	User logged in successfully	172.22.16.1:64533	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	t	\N	info		\N	2026-01-12 12:52:11.724679+03
178	2026-01-12 13:00:42.497908+03	2026-01-12 13:00:42.497908+03	\N	3	3	login	auth	User logged in successfully	[::1]:39656	PostmanRuntime/7.51.0	t	\N	info		\N	2026-01-12 13:00:42.461569+03
179	2026-01-12 16:14:23.129343+03	2026-01-12 16:14:23.129343+03	\N	3	3	login	auth	User logged in successfully	[::1]:44296	PostmanRuntime/7.51.0	t	\N	info		\N	2026-01-12 16:14:23.103853+03
180	2026-01-12 19:38:34.420539+03	2026-01-12 19:38:34.420539+03	\N	2	\N	login_failed	auth	Failed login attempt: invalid password	172.22.16.1:59048	Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36	t	\N	warning		\N	2026-01-12 19:38:34.380412+03
181	2026-01-12 19:38:39.540404+03	2026-01-12 19:38:39.540404+03	\N	2	2	login	auth	User logged in successfully	172.22.16.1:59048	Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36	t	\N	info		\N	2026-01-12 19:38:39.523599+03
182	2026-01-12 20:07:02.749472+03	2026-01-12 20:07:02.749472+03	\N	10	9	account_created	auth	User account registered	172.22.16.1:61517	Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36	t	\N	info		\N	2026-01-12 20:07:02.732074+03
183	2026-01-12 20:07:57.186301+03	2026-01-12 20:07:57.186301+03	\N	10	9	login	auth	User logged in successfully	172.22.16.1:61517	Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36	t	\N	info		\N	2026-01-12 20:07:57.158488+03
184	2026-01-12 20:09:02.81705+03	2026-01-12 20:09:02.81705+03	\N	11	10	account_created	auth	User account registered	172.22.16.1:61517	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	t	\N	info		\N	2026-01-12 20:09:02.77746+03
185	2026-01-12 20:10:10.678726+03	2026-01-12 20:10:10.678726+03	\N	11	10	login	auth	User logged in successfully	172.22.16.1:61517	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	t	\N	info		\N	2026-01-12 20:10:10.666075+03
\.


--
-- Data for Name: account_payment_gateways; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.account_payment_gateways (id, created_at, updated_at, deleted_at, account_id, payment_gateway_id, config) FROM stdin;
\.


--
-- Data for Name: accounts; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.accounts (id, created_at, updated_at, deleted_at, first_name, last_name, email, timezone_id, date_format_id, date_time_format_id, currency_id, last_ip, last_login_date, address1, address2, city, county, postal_code, is_active, is_banned, stripe_access_token, stripe_refresh_token, stripe_secret_key, stripe_publishable_key, stripe_data_raw, payment_gateway_id) FROM stdin;
2	2025-12-08 04:38:59.915039+03	2025-12-08 04:38:59.915039+03	\N	Kamau	Simon	kamausimon217@gmail.com	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	t	f	\N	\N	\N	\N	\N	\N
4	2025-12-09 20:57:42.213398+03	2025-12-09 20:57:42.213398+03	\N	Admin	kamau	topstonehelp@gmail.com	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	t	f	\N	\N	\N	\N	\N	\N
5	2025-12-10 01:27:17.964664+03	2025-12-10 01:27:17.964664+03	\N	Admin	Harsha	harshasg10@gmail.com	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	t	f	\N	\N	\N	\N	\N	\N
3	2025-12-09 05:55:26.560988+03	2025-12-11 04:41:42.725289+03	\N	symon	kamau	topstonewriters@gmail.com	1	2	1	2	\N	\N	cbd mfangano street bazaar building	floor 5 room 210	nairobi	nairobi	00100-176398	t	f	\N	\N	\N	\N	\N	\N
1	2025-12-03 00:27:40.828463+03	2025-12-11 04:59:00.887312+03	\N	Test	user	test@example.com	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	t	f	\N	\N	\N	\N	\N	\N
6	2025-12-29 13:27:28.850646+03	2025-12-29 13:27:28.850646+03	\N	Test	User	testuser@example.com	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	t	f	\N	\N	\N	\N	\N	\N
10	2026-01-12 20:07:02.638274+03	2026-01-12 20:07:02.638274+03	\N	Dorcas	Gichuru	dorcaswairuri98@gmail.com	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	t	f	\N	\N	\N	\N	\N	\N
11	2026-01-12 20:09:02.549677+03	2026-01-12 20:09:02.549677+03	\N	Dorcas	Gichuru	gichuruwairuri98@gmail.com	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	t	f	\N	\N	\N	\N	\N	\N
\.


--
-- Data for Name: attendees; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.attendees (id, created_at, updated_at, deleted_at, order_id, event_id, ticket_id, first_name, last_name, email, has_arrived, arrival_time, account_id, is_refunded, private_reference_number) FROM stdin;
\.


--
-- Data for Name: currencies; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.currencies (id, created_at, updated_at, deleted_at, code, name, symbol, is_active) FROM stdin;
1	2025-12-11 02:41:01.065589+03	2025-12-11 02:41:01.065589+03	\N	USD	US Dollar	$	t
2	2025-12-11 02:41:01.12543+03	2025-12-11 02:41:01.12543+03	\N	KSH	Kenyan Shilling	KSh	t
3	2025-12-11 02:41:01.177899+03	2025-12-11 02:41:01.177899+03	\N	EUR	Euro	€	t
4	2025-12-11 02:41:01.231058+03	2025-12-11 02:41:01.231058+03	\N	GBP	British Pound Sterling	£	t
5	2025-12-11 02:41:01.275928+03	2025-12-11 02:41:01.275928+03	\N	NGN	Nigerian Naira	₦	t
6	2025-12-11 02:41:01.334847+03	2025-12-11 02:41:01.334847+03	\N	ZAR	South African Rand	R	t
7	2025-12-11 02:41:01.401269+03	2025-12-11 02:41:01.401269+03	\N	GHS	Ghanaian Cedi	GH₵	t
8	2025-12-11 02:41:01.451771+03	2025-12-11 02:41:01.451771+03	\N	UGX	Ugandan Shilling	USh	t
9	2025-12-11 02:41:01.513539+03	2025-12-11 02:41:01.513539+03	\N	TZS	Tanzanian Shilling	TSh	t
10	2025-12-11 02:41:01.569494+03	2025-12-11 02:41:01.569494+03	\N	CAD	Canadian Dollar	C$	t
11	2025-12-11 02:41:01.623282+03	2025-12-11 02:41:01.623282+03	\N	AUD	Australian Dollar	A$	t
12	2025-12-11 02:41:01.687546+03	2025-12-11 02:41:01.687546+03	\N	INR	Indian Rupee	₹	t
13	2025-12-11 02:41:01.733328+03	2025-12-11 02:41:01.733328+03	\N	JPY	Japanese Yen	¥	t
14	2025-12-11 02:41:01.786613+03	2025-12-11 02:41:01.786613+03	\N	CNY	Chinese Yuan	¥	t
\.


--
-- Data for Name: date_formats; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.date_formats (id, created_at, updated_at, deleted_at, format, example, is_active) FROM stdin;
1	2025-12-11 02:41:01.858609+03	2025-12-11 02:41:01.858609+03	\N	YYYY-MM-DD	2024-12-25	t
2	2025-12-11 02:41:01.918008+03	2025-12-11 02:41:01.918008+03	\N	DD/MM/YYYY	25/12/2024	t
3	2025-12-11 02:41:01.972556+03	2025-12-11 02:41:01.972556+03	\N	MM/DD/YYYY	12/25/2024	t
4	2025-12-11 02:41:02.027048+03	2025-12-11 02:41:02.027048+03	\N	DD-MM-YYYY	25-12-2024	t
5	2025-12-11 02:41:02.111385+03	2025-12-11 02:41:02.111385+03	\N	MMM DD, YYYY	Dec 25, 2024	t
6	2025-12-11 02:41:02.174279+03	2025-12-11 02:41:02.174279+03	\N	DD MMM YYYY	25 Dec 2024	t
7	2025-12-11 02:41:02.266362+03	2025-12-11 02:41:02.266362+03	\N	YYYY/MM/DD	2024/12/25	t
8	2025-12-11 02:41:02.349974+03	2025-12-11 02:41:02.349974+03	\N	DD.MM.YYYY	25.12.2024	t
\.


--
-- Data for Name: date_time_formats; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.date_time_formats (id, created_at, updated_at, deleted_at, format, example, is_active) FROM stdin;
1	2025-12-11 02:41:02.420885+03	2025-12-11 02:41:02.420885+03	\N	YYYY-MM-DD HH:mm	2024-12-25 14:30	t
2	2025-12-11 02:41:02.516322+03	2025-12-11 02:41:02.516322+03	\N	DD/MM/YYYY HH:mm	25/12/2024 14:30	t
3	2025-12-11 02:41:02.610968+03	2025-12-11 02:41:02.610968+03	\N	MM/DD/YYYY hh:mm A	12/25/2024 02:30 PM	t
4	2025-12-11 02:41:02.693322+03	2025-12-11 02:41:02.693322+03	\N	DD-MM-YYYY HH:mm	25-12-2024 14:30	t
5	2025-12-11 02:41:02.781703+03	2025-12-11 02:41:02.781703+03	\N	MMM DD, YYYY hh:mm A	Dec 25, 2024 02:30 PM	t
6	2025-12-11 02:41:02.869593+03	2025-12-11 02:41:02.869593+03	\N	DD MMM YYYY HH:mm	25 Dec 2024 14:30	t
7	2025-12-11 02:41:09.528863+03	2025-12-11 02:41:09.528863+03	\N	YYYY/MM/DD HH:mm	2024/12/25 14:30	t
8	2025-12-11 02:41:09.646106+03	2025-12-11 02:41:09.646106+03	\N	DD.MM.YYYY HH:mm	25.12.2024 14:30	t
\.


--
-- Data for Name: email_verifications; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.email_verifications (id, created_at, updated_at, deleted_at, user_id, token, email, status, verified_at, expires_at, last_sent_at, resend_count, max_resends, ip_address, user_agent, issued_at) FROM stdin;
2	2025-12-08 04:39:00.097314+03	2025-12-08 06:25:54.319533+03	\N	2	e49c3550c1d997df7dddd1414aadea71	kamausimon217@gmail.com	verified	2025-12-08 06:25:54.29383+03	2025-12-09 04:39:00.074977+03	2025-12-08 04:39:00.075111+03	0	3	[::1]:38936	PostmanRuntime/7.49.1	2025-12-08 04:39:00.07512+03
1	2025-12-03 00:27:40.882023+03	2025-12-08 06:41:42.191795+03	\N	1	a0f4eb642a57c283444b6e2fb0dcdc91	test@example.com	pending	\N	2025-12-09 06:41:42.127107+03	2025-12-08 06:41:42.127107+03	1	3	[::1]:42590	PostmanRuntime/7.49.1	2025-12-03 00:27:40.881308+03
3	2025-12-09 05:55:26.773405+03	2025-12-09 05:57:44.66901+03	\N	3	c72711681eddf9e48c6018432560f21b	topstonewriters@gmail.com	verified	2025-12-09 05:57:44.651308+03	2025-12-10 05:55:26.728189+03	2025-12-09 05:55:26.728589+03	0	3	[::1]:59792	PostmanRuntime/7.49.1	2025-12-09 05:55:26.728593+03
5	2025-12-09 06:31:29.537493+03	2025-12-09 06:37:40.64613+03	\N	3	f20eb800e9880c7104ae6c2aec1cb23fad70737f8f02f2316471a5d0f4fcc8a6	topstonewriters@gmail.com	pending	2025-12-09 06:37:40.645992+03	2025-12-10 06:31:29.496302+03	2025-12-09 06:31:29.496306+03	0	3	[::1]	PostmanRuntime/7.49.1	2025-12-09 06:31:29.496305+03
6	2025-12-09 20:57:42.321372+03	2025-12-09 20:58:20.053475+03	\N	4	684d5fada3c4aa70dadfca8f755a11ae	topstonehelp@gmail.com	verified	2025-12-09 20:58:20.030324+03	2025-12-10 20:57:42.293362+03	2025-12-09 20:57:42.293984+03	0	3	[::1]:59370	PostmanRuntime/7.49.1	2025-12-09 20:57:42.293987+03
7	2025-12-10 01:27:12.483667+03	2025-12-10 01:27:50.295834+03	\N	5	984d782d10e45b312e1820080d817a23	harshasg10@gmail.com	verified	2025-12-10 01:27:50.278481+03	2025-12-11 01:27:12.454503+03	2025-12-10 01:27:12.454579+03	0	3	[::1]:47204	PostmanRuntime/7.49.1	2025-12-10 01:27:12.454586+03
8	2025-12-29 13:27:28.953734+03	2025-12-29 13:27:28.953734+03	\N	6	66939702469fd8b291dcea4e68ca3144	testuser@example.com	pending	\N	2025-12-30 13:27:28.932995+03	2025-12-29 13:27:28.93308+03	0	3	[::1]:34730	curl/8.5.0	2025-12-29 13:27:28.933086+03
9	2026-01-12 20:07:02.707072+03	2026-01-12 20:07:02.707072+03	\N	9	f521fe2b962e97a3eb0f3b82fb79bbbc	dorcaswairuri98@gmail.com	pending	\N	2026-01-13 20:07:02.690239+03	2026-01-12 20:07:02.690255+03	0	3	172.22.16.1:61517	Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36	2026-01-12 20:07:02.690261+03
10	2026-01-12 20:09:02.76735+03	2026-01-12 20:09:02.76735+03	\N	10	17147ba445cafeb1956768f311d7735f	gichuruwairuri98@gmail.com	pending	\N	2026-01-13 20:09:02.732633+03	2026-01-12 20:09:02.732636+03	0	3	172.22.16.1:61517	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	2026-01-12 20:09:02.732639+03
\.


--
-- Data for Name: event_images; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.event_images (id, created_at, updated_at, deleted_at, image_path, event_id, account_id, user_id) FROM stdin;
4	2025-12-10 05:54:15.714088+03	2025-12-10 05:54:15.714088+03	2025-12-10 05:58:13.04401+03	uploads/events/event_2_1765335255.jpg	2	3	3
5	2025-12-10 06:01:19.873893+03	2025-12-10 06:01:19.873893+03	\N	uploads/events/event_2_1765335679.jpg	2	3	3
11	2026-01-12 16:40:20.195449+03	2026-01-12 16:40:20.195449+03	\N	https://aws-ticketing-bucket-ksm.s3.us-east-1.amazonaws.com/events/ebd0fbec-12f9-411a-8fc3-d4acd525d4cb.jpg	4	3	3
\.


--
-- Data for Name: event_metrics; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.event_metrics (id, created_at, updated_at, deleted_at, event_id, date, hour, page_views, unique_visitors, bounce_rate, avg_time_on_page, add_to_cart, checkout_start, checkout_complete, conversion_rate, gross_revenue, net_revenue, platform_fees, payment_fees, refund_amount, tickets_sold, tickets_refunded, tickets_checked_in, inventory_remaining, promo_code_uses, promo_discount, top_countries, top_cities, mobile_percent, desktop_percent, app_percent) FROM stdin;
\.


--
-- Data for Name: event_stats; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.event_stats (id, created_at, updated_at, deleted_at, date, day, hour, views, unique_views, tickets_sold, sales_volume, organizer_fee_volume, event_id, add_to_cart_count, check_out_start, conversion_rate, gross_revenue, net_revenue, platform_fees, payment_fees, average_time_on_page, bounce_rate, granularity) FROM stdin;
\.


--
-- Data for Name: event_venues; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.event_venues (id, created_at, updated_at, deleted_at, venue_id, event_id, venue_role, setup_time, event_time, cleanup_time) FROM stdin;
\.


--
-- Data for Name: events; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.events (id, created_at, updated_at, deleted_at, title, location, bg_type, bg_color, description, start_date, end_date, on_sale_date, organizer_id, account_id, sales_volume, organizer_fees_volume, organizer_fee_fixed, organizer_fee_percentage, currency, location_address, location_address_line, location_country, pre_order_message_display, post_order_message_display, is_live, barcode_type, is_barcode_enabled, ticket_border_color, ticket_bg_color, ticket_text_color, ticket_sub_text_color, enable_offline_payment, max_capacity, status, category, tags, min_age, is_private) FROM stdin;
2	2025-12-10 05:18:12.437869+03	2025-12-10 06:05:41.551164+03	\N	kikuyu love songs	kiambu bamboo resort	color	blue	Welcome we enjoy some mugithi together	2025-12-30 15:00:00+03	2025-12-31 10:00:00+03	2025-12-17 03:00:00+03	1	3	0	0	1	3.5	ksh	kiambu bamboo resort along limuru road	127T limuru	Kenya	We welcome you to our event, be the first to get your ticket before they sell out	Thank you for your order	t	qr code	t	green	red	black	blue	t	5000	live	music	kikuyu, mugithi	18	f
4	2026-01-10 21:55:30.330855+03	2026-01-12 13:05:31.318358+03	\N	Buy the dev a coffee	online	color	blue	Let's explore together	2026-12-30 15:00:00+03	2026-12-31 10:00:00+03	2026-01-11 03:00:00+03	1	3	0	0	1	3.5	ksh	online	\N	Kenya	We welcome you to our event, be the first to get your ticket before they sell out	Thank you for your order	t	qr code	t	green	red	black	blue	t	200	live	art	event	1	f
\.


--
-- Data for Name: login_history; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.login_history (id, created_at, updated_at, deleted_at, account_id, user_id, ip_address, user_agent, location, device, browser, success, fail_reason, login_at, logout_at, session_duration) FROM stdin;
1	2025-12-03 03:58:50.504517+03	2025-12-03 03:58:50.504517+03	\N	1	1	[::1]:46836	curl/8.5.0	\N	\N	\N	t	\N	2025-12-03 03:58:50.480425+03	\N	\N
2	2025-12-03 04:12:15.76083+03	2025-12-03 04:12:15.76083+03	\N	1	1	[::1]:59284	curl/8.5.0	\N	\N	\N	t	\N	2025-12-03 04:12:15.743959+03	\N	\N
3	2025-12-03 04:16:23.034917+03	2025-12-03 04:16:23.034917+03	\N	1	1	[::1]:42012	curl/8.5.0	\N	\N	\N	t	\N	2025-12-03 04:16:23.018742+03	\N	\N
4	2025-12-03 04:17:12.887846+03	2025-12-03 04:17:12.887846+03	\N	1	1	[::1]:38064	curl/8.5.0	\N	\N	\N	t	\N	2025-12-03 04:17:12.87005+03	\N	\N
5	2025-12-03 04:21:01.255392+03	2025-12-03 04:21:01.255392+03	\N	1	1	[::1]:59652	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-03 04:21:01.232401+03	\N	\N
6	2025-12-08 03:13:07.640443+03	2025-12-08 03:13:07.640443+03	\N	1	1	[::1]:52608	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-08 03:13:07.620376+03	\N	\N
7	2025-12-08 04:40:58.69127+03	2025-12-08 04:40:58.69127+03	\N	2	2	[::1]:40272	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-08 04:40:58.65751+03	\N	\N
8	2025-12-08 04:41:32.417873+03	2025-12-08 04:41:32.417873+03	\N	2	2	[::1]:40272	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-08 04:41:32.404024+03	\N	\N
9	2025-12-08 06:16:34.04907+03	2025-12-08 06:16:34.04907+03	\N	2	\N	[::1]:44702	PostmanRuntime/7.49.1	\N	\N	\N	f	invalid password	2025-12-08 06:16:34.025133+03	\N	\N
10	2025-12-08 06:16:43.408487+03	2025-12-08 06:16:43.408487+03	\N	2	2	[::1]:44702	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-08 06:16:43.39125+03	\N	\N
11	2025-12-08 06:17:52.551555+03	2025-12-08 06:17:52.551555+03	\N	2	\N	[::1]:44702	PostmanRuntime/7.49.1	\N	\N	\N	f	invalid password	2025-12-08 06:17:52.537669+03	\N	\N
12	2025-12-08 06:17:56.101604+03	2025-12-08 06:17:56.101604+03	\N	2	2	[::1]:44702	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-08 06:17:56.060926+03	\N	\N
13	2025-12-08 06:44:59.724179+03	2025-12-08 06:44:59.724179+03	\N	2	2	[::1]:52274	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-08 06:44:59.707737+03	\N	\N
14	2025-12-09 01:49:59.047629+03	2025-12-09 01:49:59.047629+03	\N	2	2	[::1]:48172	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-09 01:49:58.994745+03	\N	\N
15	2025-12-09 02:27:18.739462+03	2025-12-09 02:27:18.739462+03	\N	2	2	[::1]:55558	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-09 02:27:18.585668+03	\N	\N
16	2025-12-09 02:43:52.975202+03	2025-12-09 02:43:52.975202+03	\N	2	2	[::1]:41086	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-09 02:43:52.920084+03	\N	\N
17	2025-12-09 06:00:09.842526+03	2025-12-09 06:00:09.842526+03	\N	3	3	[::1]:42464	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-09 06:00:09.805187+03	\N	\N
18	2025-12-09 07:08:22.256804+03	2025-12-09 07:08:22.256804+03	\N	3	3	[::1]:57596	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-09 07:08:22.238344+03	\N	\N
19	2025-12-09 20:34:29.880957+03	2025-12-09 20:34:29.880957+03	\N	3	3	[::1]:33654	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-09 20:34:29.859947+03	\N	\N
20	2025-12-09 20:58:47.487924+03	2025-12-09 20:58:47.487924+03	\N	4	4	[::1]:59370	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-09 20:58:47.466262+03	\N	\N
21	2025-12-09 21:06:45.311053+03	2025-12-09 21:06:45.311053+03	\N	4	4	[::1]:51588	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-09 21:06:45.29413+03	\N	\N
22	2025-12-09 21:08:28.565272+03	2025-12-09 21:08:28.565272+03	\N	4	4	[::1]:56596	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-09 21:08:28.536806+03	\N	\N
23	2025-12-09 21:08:45.409412+03	2025-12-09 21:08:45.409412+03	\N	4	4	[::1]:56596	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-09 21:08:45.394593+03	\N	\N
24	2025-12-09 21:10:15.647419+03	2025-12-09 21:10:15.647419+03	\N	4	4	[::1]:40480	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-09 21:10:15.626107+03	\N	\N
25	2025-12-09 21:12:11.149057+03	2025-12-09 21:12:11.149057+03	\N	4	4	[::1]:40934	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-09 21:12:11.134623+03	\N	\N
26	2025-12-09 21:23:20.796156+03	2025-12-09 21:23:20.796156+03	\N	3	3	[::1]:56326	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-09 21:23:20.77535+03	\N	\N
27	2025-12-09 21:23:37.26603+03	2025-12-09 21:23:37.26603+03	\N	4	4	[::1]:56326	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-09 21:23:37.234208+03	\N	\N
28	2025-12-09 23:42:19.999415+03	2025-12-09 23:42:19.999415+03	\N	4	4	[::1]:55226	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-09 23:42:19.979691+03	\N	\N
29	2025-12-10 00:38:45.343128+03	2025-12-10 00:38:45.343128+03	\N	3	3	[::1]:48936	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-10 00:38:45.298639+03	\N	\N
30	2025-12-10 00:51:35.459207+03	2025-12-10 00:51:35.459207+03	\N	4	4	[::1]:49796	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-10 00:51:35.443467+03	\N	\N
31	2025-12-10 01:28:13.743538+03	2025-12-10 01:28:13.743538+03	\N	5	5	[::1]:47204	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-10 01:28:13.725696+03	\N	\N
32	2025-12-10 01:28:33.282715+03	2025-12-10 01:28:33.282715+03	\N	4	4	[::1]:47204	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-10 01:28:33.259859+03	\N	\N
33	2025-12-10 01:44:11.450443+03	2025-12-10 01:44:11.450443+03	\N	3	3	[::1]:40868	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-10 01:44:11.426831+03	\N	\N
34	2025-12-10 04:44:03.221561+03	2025-12-10 04:44:03.221561+03	\N	3	3	[::1]:49102	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-10 04:44:03.207233+03	\N	\N
35	2025-12-10 05:45:43.398794+03	2025-12-10 05:45:43.398794+03	\N	3	3	[::1]:45984	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-10 05:45:43.295918+03	\N	\N
36	2025-12-10 05:58:05.639004+03	2025-12-10 05:58:05.639004+03	\N	3	3	[::1]:36080	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-10 05:58:05.621559+03	\N	\N
37	2025-12-10 07:02:23.531623+03	2025-12-10 07:02:23.531623+03	\N	3	3	[::1]:38390	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-10 07:02:23.517185+03	\N	\N
38	2025-12-10 22:09:13.805135+03	2025-12-10 22:09:13.805135+03	\N	3	3	[::1]:42504	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-10 22:09:13.78078+03	\N	\N
39	2025-12-11 00:12:50.229269+03	2025-12-11 00:12:50.229269+03	\N	3	3	[::1]:53738	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-11 00:12:50.213018+03	\N	\N
40	2025-12-11 00:17:03.733012+03	2025-12-11 00:17:03.733012+03	\N	1	1	[::1]:51510	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-11 00:17:03.717669+03	\N	\N
41	2025-12-11 00:18:20.861625+03	2025-12-11 00:18:20.861625+03	\N	1	1	[::1]:51510	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-11 00:18:20.844037+03	\N	\N
42	2025-12-11 01:55:23.553495+03	2025-12-11 01:55:23.553495+03	\N	3	3	[::1]:49560	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-11 01:55:23.529546+03	\N	\N
43	2025-12-11 02:52:10.193126+03	2025-12-11 02:52:10.193126+03	\N	3	3	[::1]:51162	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-11 02:52:10.174369+03	\N	\N
44	2025-12-11 04:38:12.985055+03	2025-12-11 04:38:12.985055+03	\N	3	3	[::1]:45492	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-11 04:38:12.971271+03	\N	\N
45	2025-12-11 04:50:00.506822+03	2025-12-11 04:50:00.506822+03	\N	3	\N	[::1]:56360	PostmanRuntime/7.49.1	\N	\N	\N	f	invalid password	2025-12-11 04:50:00.486778+03	\N	\N
46	2025-12-11 04:50:18.496882+03	2025-12-11 04:50:18.496882+03	\N	3	3	[::1]:56360	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-11 04:50:18.470868+03	\N	\N
47	2025-12-11 04:51:51.903165+03	2025-12-11 04:51:51.903165+03	\N	3	3	[::1]:56360	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-11 04:51:51.884788+03	\N	\N
48	2025-12-11 04:52:00.817351+03	2025-12-11 04:52:00.817351+03	\N	4	4	[::1]:56360	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-11 04:52:00.799809+03	\N	\N
49	2025-12-11 04:59:16.322989+03	2025-12-11 04:59:16.322989+03	\N	3	3	[::1]:60092	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-11 04:59:16.273767+03	\N	\N
50	2025-12-11 05:24:23.565591+03	2025-12-11 05:24:23.565591+03	\N	2	2	[::1]:51290	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-11 05:24:23.543679+03	\N	\N
51	2025-12-11 05:26:35.621433+03	2025-12-11 05:26:35.621433+03	\N	3	3	[::1]:41310	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-11 05:26:35.602929+03	\N	\N
52	2025-12-11 05:31:34.674671+03	2025-12-11 05:31:34.674671+03	\N	2	2	[::1]:37622	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-11 05:31:34.653184+03	\N	\N
53	2025-12-12 01:49:24.651888+03	2025-12-12 01:49:24.651888+03	\N	3	3	[::1]:41418	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-12 01:49:24.626188+03	\N	\N
54	2025-12-12 05:24:16.306437+03	2025-12-12 05:24:16.306437+03	\N	3	3	[::1]:48578	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-12 05:24:16.27644+03	\N	\N
55	2025-12-12 05:28:05.106287+03	2025-12-12 05:28:05.106287+03	\N	2	2	[::1]:56138	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-12 05:28:05.087397+03	\N	\N
56	2025-12-12 06:31:49.215566+03	2025-12-12 06:31:49.215566+03	\N	2	2	[::1]:40782	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-12 06:31:49.1948+03	\N	\N
57	2025-12-15 06:23:33.806608+03	2025-12-15 06:23:33.806608+03	\N	2	2	[::1]:59658	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-15 06:23:33.794087+03	\N	\N
58	2025-12-16 04:22:50.248492+03	2025-12-16 04:22:50.248492+03	\N	2	2	[::1]:50112	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-16 04:22:50.228601+03	\N	\N
59	2025-12-16 05:09:34.078062+03	2025-12-16 05:09:34.078062+03	\N	2	2	[::1]:48612	PostmanRuntime/7.49.1	\N	\N	\N	t	\N	2025-12-16 05:09:33.87971+03	\N	\N
60	2025-12-16 05:32:29.867817+03	2025-12-16 05:32:29.867817+03	\N	2	2	[::1]:34744	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-16 05:32:29.827795+03	\N	\N
61	2025-12-16 06:24:32.602456+03	2025-12-16 06:24:32.602456+03	\N	2	2	[::1]:59036	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-16 06:24:32.582969+03	\N	\N
62	2025-12-16 07:53:56.15494+03	2025-12-16 07:53:56.15494+03	\N	2	2	[::1]:49694	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-16 07:53:56.138735+03	\N	\N
63	2025-12-16 08:06:54.462482+03	2025-12-16 08:06:54.462482+03	\N	4	4	[::1]:50524	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-16 08:06:54.434927+03	\N	\N
64	2025-12-22 01:50:14.142547+03	2025-12-22 01:50:14.142547+03	\N	2	2	[::1]:37822	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-22 01:50:14.128449+03	\N	\N
65	2025-12-22 01:53:02.416655+03	2025-12-22 01:53:02.416655+03	\N	4	4	[::1]:35580	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-22 01:53:02.402921+03	\N	\N
66	2025-12-22 03:34:09.077276+03	2025-12-22 03:34:09.077276+03	\N	4	4	[::1]:33074	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-22 03:34:09.056334+03	\N	\N
67	2025-12-22 05:24:05.78092+03	2025-12-22 05:24:05.78092+03	\N	3	3	[::1]:55030	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-22 05:24:05.753181+03	\N	\N
68	2025-12-22 05:47:00.005765+03	2025-12-22 05:47:00.005765+03	\N	3	3	[::1]:57160	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-22 05:46:59.981575+03	\N	\N
69	2025-12-22 06:12:48.091711+03	2025-12-22 06:12:48.091711+03	\N	2	2	[::1]:50848	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-22 06:12:48.073456+03	\N	\N
70	2025-12-22 07:09:16.767076+03	2025-12-22 07:09:16.767076+03	\N	4	4	[::1]:50940	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-22 07:09:16.742428+03	\N	\N
71	2025-12-22 07:11:35.201377+03	2025-12-22 07:11:35.201377+03	\N	2	2	[::1]:50940	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-22 07:11:35.181132+03	\N	\N
72	2025-12-23 13:06:43.494804+03	2025-12-23 13:06:43.494804+03	\N	2	2	[::1]:38442	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-23 13:06:43.469483+03	\N	\N
73	2025-12-23 14:18:12.31358+03	2025-12-23 14:18:12.31358+03	\N	2	2	[::1]:39316	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-23 14:18:12.299843+03	\N	\N
74	2025-12-23 14:57:46.782221+03	2025-12-23 14:57:46.782221+03	\N	3	3	[::1]:50148	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-23 14:57:46.763407+03	\N	\N
75	2025-12-24 01:26:17.006829+03	2025-12-24 01:26:17.006829+03	\N	3	3	[::1]:54464	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-24 01:26:16.978596+03	\N	\N
76	2025-12-24 02:34:16.707541+03	2025-12-24 02:34:16.707541+03	\N	3	3	[::1]:34136	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-24 02:34:16.687414+03	\N	\N
77	2025-12-24 03:55:12.098762+03	2025-12-24 03:55:12.098762+03	\N	3	3	[::1]:35634	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-24 03:55:12.076081+03	\N	\N
78	2025-12-24 04:31:38.856346+03	2025-12-24 04:31:38.856346+03	\N	3	3	[::1]:51086	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-24 04:31:38.82282+03	\N	\N
79	2025-12-24 09:10:15.601265+03	2025-12-24 09:10:15.601265+03	\N	3	3	[::1]:56662	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-24 09:10:21.630823+03	\N	\N
80	2025-12-24 10:13:39.690498+03	2025-12-24 10:13:39.690498+03	\N	3	3	[::1]:53098	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-24 10:13:39.655664+03	\N	\N
81	2025-12-24 10:50:10.201886+03	2025-12-24 10:50:10.201886+03	\N	4	4	[::1]:52786	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-24 10:50:10.176628+03	\N	\N
82	2025-12-24 10:50:52.722853+03	2025-12-24 10:50:52.722853+03	\N	3	3	[::1]:52786	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-24 10:50:52.644344+03	\N	\N
83	2025-12-24 11:14:44.408538+03	2025-12-24 11:14:44.408538+03	\N	3	3	[::1]:58690	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-24 11:14:44.384596+03	\N	\N
84	2025-12-24 11:17:58.69388+03	2025-12-24 11:17:58.69388+03	\N	3	3	[::1]:53768	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-24 11:17:58.668852+03	\N	\N
85	2025-12-24 11:30:46.431857+03	2025-12-24 11:30:46.431857+03	\N	3	3	[::1]:35838	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-24 11:30:46.413013+03	\N	\N
86	2025-12-24 12:11:11.040231+03	2025-12-24 12:11:11.040231+03	\N	3	3	[::1]:49494	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-24 12:11:11.015941+03	\N	\N
87	2025-12-28 06:34:46.665369+03	2025-12-28 06:34:46.665369+03	\N	2	2	[::1]:46380	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-28 06:34:46.642863+03	\N	\N
88	2025-12-28 09:13:49.641238+03	2025-12-28 09:13:49.641238+03	\N	2	2	[::1]:60496	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-28 09:13:49.604949+03	\N	\N
89	2025-12-28 09:24:35.023995+03	2025-12-28 09:24:35.023995+03	\N	2	2	[::1]:33700	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-28 09:24:35.011235+03	\N	\N
90	2025-12-28 09:29:13.624901+03	2025-12-28 09:29:13.624901+03	\N	2	2	[::1]:48924	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-28 09:29:13.612036+03	\N	\N
91	2025-12-28 09:46:37.081029+03	2025-12-28 09:46:37.081029+03	\N	2	2	[::1]:50236	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-28 09:46:37.061075+03	\N	\N
92	2025-12-28 11:10:38.785805+03	2025-12-28 11:10:38.785805+03	\N	2	2	[::1]:37882	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-28 11:10:38.758852+03	\N	\N
93	2025-12-28 18:19:57.058144+03	2025-12-28 18:19:57.058144+03	\N	2	2	[::1]:44662	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-28 18:19:57.02425+03	\N	\N
94	2025-12-28 18:48:21.388637+03	2025-12-28 18:48:21.388637+03	\N	2	2	[::1]:47434	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-28 18:48:21.369653+03	\N	\N
95	2025-12-29 13:19:29.257747+03	2025-12-29 13:19:29.257747+03	\N	2	2	[::1]:40062	Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36	\N	\N	\N	t	\N	2025-12-29 13:19:29.240518+03	\N	\N
96	2025-12-29 13:23:41.280571+03	2025-12-29 13:23:41.280571+03	\N	2	2	[::1]:40062	Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36	\N	\N	\N	t	\N	2025-12-29 13:23:41.264108+03	\N	\N
97	2025-12-29 13:23:54.04608+03	2025-12-29 13:23:54.04608+03	\N	2	2	[::1]:40062	Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36	\N	\N	\N	t	\N	2025-12-29 13:23:54.027184+03	\N	\N
98	2025-12-29 13:24:37.387631+03	2025-12-29 13:24:37.387631+03	\N	2	2	[::1]:40062	Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36	\N	\N	\N	t	\N	2025-12-29 13:24:37.374755+03	\N	\N
99	2025-12-29 13:25:49.190001+03	2025-12-29 13:25:49.190001+03	\N	2	2	[::1]:40062	Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36	\N	\N	\N	t	\N	2025-12-29 13:25:49.170232+03	\N	\N
100	2025-12-29 13:26:04.310605+03	2025-12-29 13:26:04.310605+03	\N	2	2	[::1]:40062	Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36	\N	\N	\N	t	\N	2025-12-29 13:26:04.290416+03	\N	\N
101	2025-12-29 13:26:39.504533+03	2025-12-29 13:26:39.504533+03	\N	2	2	[::1]:40062	Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36	\N	\N	\N	t	\N	2025-12-29 13:26:39.491909+03	\N	\N
102	2025-12-29 13:27:06.807485+03	2025-12-29 13:27:06.807485+03	\N	1	\N	[::1]:45610	curl/8.5.0	\N	\N	\N	f	invalid password	2025-12-29 13:27:06.788193+03	\N	\N
103	2025-12-29 13:27:39.678078+03	2025-12-29 13:27:39.678078+03	\N	6	6	[::1]:43206	curl/8.5.0	\N	\N	\N	t	\N	2025-12-29 13:27:39.660739+03	\N	\N
104	2025-12-29 13:30:26.897649+03	2025-12-29 13:30:26.897649+03	\N	2	2	[::1]:40062	Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36	\N	\N	\N	t	\N	2025-12-29 13:30:26.885014+03	\N	\N
105	2025-12-29 13:30:33.850723+03	2025-12-29 13:30:33.850723+03	\N	2	2	[::1]:40062	Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36	\N	\N	\N	t	\N	2025-12-29 13:30:33.815531+03	\N	\N
106	2025-12-29 13:31:26.749538+03	2025-12-29 13:31:26.749538+03	\N	2	2	[::1]:40062	Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36	\N	\N	\N	t	\N	2025-12-29 13:31:26.733817+03	\N	\N
107	2025-12-29 15:13:16.644667+03	2025-12-29 15:13:16.644667+03	\N	2	2	[::1]:56814	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	\N	\N	\N	t	\N	2025-12-29 15:13:16.631289+03	\N	\N
108	2025-12-29 15:21:35.645416+03	2025-12-29 15:21:35.645416+03	\N	2	2	[::1]:49852	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	\N	\N	\N	t	\N	2025-12-29 15:21:35.56812+03	\N	\N
109	2025-12-29 15:39:36.095641+03	2025-12-29 15:39:36.095641+03	\N	2	2	[::1]:40974	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Code/1.107.1 Chrome/142.0.7444.175 Electron/39.2.3 Safari/537.36	\N	\N	\N	t	\N	2025-12-29 15:39:36.082887+03	\N	\N
110	2025-12-29 16:20:56.171888+03	2025-12-29 16:20:56.171888+03	\N	2	2	[::1]:57028	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2025-12-29 16:20:56.145321+03	\N	\N
111	2026-01-08 12:09:03.443+03	2026-01-08 12:09:03.443+03	\N	2	2	[::1]:47256	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2026-01-08 12:09:03.419728+03	\N	\N
112	2026-01-08 12:55:07.107826+03	2026-01-08 12:55:07.107826+03	\N	4	4	[::1]:36732	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2026-01-08 12:55:07.075853+03	\N	\N
113	2026-01-08 13:44:30.126364+03	2026-01-08 13:44:30.126364+03	\N	2	2	[::1]:44190	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2026-01-08 13:44:30.108547+03	\N	\N
114	2026-01-08 13:47:27.891859+03	2026-01-08 13:47:27.891859+03	\N	4	4	[::1]:32982	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2026-01-08 13:47:27.869893+03	\N	\N
115	2026-01-08 19:03:32.738652+03	2026-01-08 19:03:32.738652+03	\N	4	4	[::1]:45670	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2026-01-08 19:03:32.701928+03	\N	\N
116	2026-01-08 19:03:44.363234+03	2026-01-08 19:03:44.363234+03	\N	2	2	[::1]:45670	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2026-01-08 19:03:44.350597+03	\N	\N
117	2026-01-08 19:10:44.265655+03	2026-01-08 19:10:44.265655+03	\N	2	2	[::1]:51714	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2026-01-08 19:10:44.228763+03	\N	\N
118	2026-01-08 20:13:08.678177+03	2026-01-08 20:13:08.678177+03	\N	2	2	[::1]:41398	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2026-01-08 20:13:08.665245+03	\N	\N
119	2026-01-08 21:16:46.663898+03	2026-01-08 21:16:46.663898+03	\N	2	2	[::1]:58828	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2026-01-08 21:16:46.643676+03	\N	\N
120	2026-01-09 09:31:22.686422+03	2026-01-09 09:31:22.686422+03	\N	4	4	[::1]:60336	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2026-01-09 09:31:22.673291+03	\N	\N
121	2026-01-09 09:51:46.76879+03	2026-01-09 09:51:46.76879+03	\N	4	4	[::1]:33748	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2026-01-09 09:51:46.755109+03	\N	\N
122	2026-01-09 10:58:36.916032+03	2026-01-09 10:58:36.916032+03	\N	4	4	[::1]:34030	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2026-01-09 10:58:36.890043+03	\N	\N
123	2026-01-09 12:40:31.413991+03	2026-01-09 12:40:31.413991+03	\N	4	4	[::1]:60414	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2026-01-09 12:40:31.382975+03	\N	\N
124	2026-01-10 18:13:59.198621+03	2026-01-10 18:13:59.198621+03	\N	2	2	[::1]:55768	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	\N	\N	\N	t	\N	2026-01-10 18:13:59.181827+03	\N	\N
125	2026-01-10 18:26:22.349786+03	2026-01-10 18:26:22.349786+03	\N	2	2	172.22.16.1:51491	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	\N	\N	\N	t	\N	2026-01-10 18:26:22.329193+03	\N	\N
126	2026-01-10 21:24:30.944481+03	2026-01-10 21:24:30.944481+03	\N	3	3	[::1]:38674	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2026-01-10 21:24:30.933191+03	\N	\N
127	2026-01-10 22:25:40.608552+03	2026-01-10 22:25:40.608552+03	\N	2	2	172.22.16.1:58337	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	\N	\N	\N	t	\N	2026-01-10 22:25:40.596598+03	\N	\N
128	2026-01-10 22:29:03.967894+03	2026-01-10 22:29:03.967894+03	\N	3	3	[::1]:52238	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2026-01-10 22:29:03.938975+03	\N	\N
129	2026-01-10 22:41:09.239018+03	2026-01-10 22:41:09.239018+03	\N	2	2	172.22.16.1:58880	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	\N	\N	\N	t	\N	2026-01-10 22:41:09.222465+03	\N	\N
130	2026-01-10 22:48:15.885054+03	2026-01-10 22:48:15.885054+03	\N	2	2	172.22.16.1:58880	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	\N	\N	\N	t	\N	2026-01-10 22:48:15.871096+03	\N	\N
131	2026-01-10 23:01:02.304129+03	2026-01-10 23:01:02.304129+03	\N	2	2	172.22.16.1:59351	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	\N	\N	\N	t	\N	2026-01-10 23:01:02.280899+03	\N	\N
132	2026-01-12 12:39:08.904478+03	2026-01-12 12:39:08.904478+03	\N	2	2	172.22.16.1:64336	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	\N	\N	\N	t	\N	2026-01-12 12:39:08.890334+03	\N	\N
133	2026-01-12 12:52:11.782155+03	2026-01-12 12:52:11.782155+03	\N	2	2	172.22.16.1:64533	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	\N	\N	\N	t	\N	2026-01-12 12:52:11.768934+03	\N	\N
134	2026-01-12 13:00:42.550313+03	2026-01-12 13:00:42.550313+03	\N	3	3	[::1]:39656	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2026-01-12 13:00:42.528495+03	\N	\N
135	2026-01-12 16:14:23.248839+03	2026-01-12 16:14:23.248839+03	\N	3	3	[::1]:44296	PostmanRuntime/7.51.0	\N	\N	\N	t	\N	2026-01-12 16:14:23.229033+03	\N	\N
136	2026-01-12 19:38:34.553868+03	2026-01-12 19:38:34.553868+03	\N	2	\N	172.22.16.1:59048	Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36	\N	\N	\N	f	invalid password	2026-01-12 19:38:34.537495+03	\N	\N
137	2026-01-12 19:38:39.577271+03	2026-01-12 19:38:39.577271+03	\N	2	2	172.22.16.1:59048	Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36	\N	\N	\N	t	\N	2026-01-12 19:38:39.563304+03	\N	\N
138	2026-01-12 20:07:57.229949+03	2026-01-12 20:07:57.229949+03	\N	10	9	172.22.16.1:61517	Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Mobile Safari/537.36	\N	\N	\N	t	\N	2026-01-12 20:07:57.204946+03	\N	\N
139	2026-01-12 20:10:10.705794+03	2026-01-12 20:10:10.705794+03	\N	11	10	172.22.16.1:61517	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36	\N	\N	\N	t	\N	2026-01-12 20:10:10.690401+03	\N	\N
\.


--
-- Data for Name: notification_preferences; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.notification_preferences (id, created_at, updated_at, deleted_at, account_id, email_notifications, sms_notifications, push_notifications, event_updates, payment_notifications, security_alerts, marketing_emails) FROM stdin;
1	2025-12-11 03:09:13.286909+03	2025-12-11 03:09:13.286909+03	\N	3	t	f	t	t	t	t	f
\.


--
-- Data for Name: order_items; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.order_items (id, created_at, updated_at, deleted_at, order_id, ticket_class_id, quantity, unit_price, total_price, discount, promo_code_used) FROM stdin;
2	2025-12-12 06:08:45.531087+03	2025-12-12 06:08:45.531087+03	\N	2	2	1	200	200	\N	\N
6	2025-12-16 06:55:16.203495+03	2025-12-16 06:55:16.203495+03	\N	6	1	2	100	200	\N	\N
7	2025-12-16 07:08:05.745669+03	2025-12-16 07:08:05.745669+03	\N	7	1	2	100	200	\N	\N
9	2025-12-16 07:23:40.55387+03	2025-12-16 07:23:40.55387+03	\N	9	1	1	100	100	\N	\N
8	2025-12-16 07:13:52.571185+03	2025-12-16 07:13:52.571185+03	\N	8	1	1	100	100	\N	\N
3	2025-12-16 06:30:02.347164+03	2025-12-16 06:30:02.347164+03	\N	3	1	1	100	100	\N	\N
1	2025-12-12 05:55:23.37848+03	2025-12-12 05:55:23.37848+03	\N	1	1	2	100	200	\N	\N
4	2025-12-16 06:39:29.480503+03	2025-12-16 06:39:29.480503+03	\N	4	1	2	100	200	\N	\N
5	2025-12-16 06:52:15.882029+03	2025-12-16 06:52:15.882029+03	\N	5	1	2	100	200	\N	\N
\.


--
-- Data for Name: orders; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.orders (id, created_at, updated_at, deleted_at, account_id, first_name, last_name, email, business_name, business_tax_number, business_address_line, ticket_pdf_path, order_preference, transaction_id, discount, booking_fee, organizer_booking_fee, order_date, notes, is_deleted, is_cancelled, is_partially_refunded, amount, amount_refunded, event_id, payment_gateway_id, is_payment_received, is_business, tax_amount, status, payment_status, total_amount, currency, completed_at, cancelled_at, refunded_at) FROM stdin;
6	2025-12-16 06:55:16.185186+03	2025-12-16 06:55:47.693255+03	\N	2	Kamau	simon	kamausimon217@gmail.com	\N	\N	\N	\N		\N	\N	6	10	2025-12-16 06:55:16.185142+03	\N	f	f	f	200	\N	2	\N	t	f	32	fulfilled	completed	0	KSH	2025-12-16 06:55:47.655589+03	\N	\N
7	2025-12-16 07:08:05.67767+03	2025-12-16 07:08:39.862352+03	\N	2	Kamau	simon	kamausimon217@gmail.com	\N	\N	\N	\N		\N	\N	6	10	2025-12-16 07:08:05.677581+03	\N	f	f	f	200	\N	2	\N	t	f	32	fulfilled	completed	0	KSH	2025-12-16 07:08:39.724294+03	\N	\N
9	2025-12-16 07:23:40.434153+03	2025-12-16 07:24:17.982261+03	\N	2	Kamau	simon	kamausimon217@gmail.com	\N	\N	\N	\N		\N	\N	3	5	2025-12-16 07:23:40.434093+03	\N	f	f	f	100	\N	2	\N	t	f	16	fulfilled	completed	0	KSH	2025-12-16 07:24:17.826993+03	\N	\N
2	2025-12-12 06:08:45.375779+03	2025-12-22 02:12:45.991743+03	\N	2	Kamau	simon	kamausimon217@gmail.com	\N	\N	\N	\N		\N	\N	6	10	2025-12-12 06:08:45.375717+03	\N	f	f	f	200	\N	2	\N	f	f	32	cancelled	pending	0	KSH	\N	\N	\N
8	2025-12-16 07:13:52.561358+03	2025-12-22 02:15:00.449813+03	\N	2	Kamau	simon	kamausimon217@gmail.com	\N	\N	\N	\N		\N	\N	3	5	2025-12-16 07:13:52.561277+03	\N	f	t	f	100	\N	2	\N	f	f	16	cancelled	pending	0	KSH	\N	2025-12-22 02:15:00.449686+03	\N
3	2025-12-16 06:30:02.272685+03	2025-12-22 02:42:28.021275+03	\N	2	Kamau	simon	kamausimon217@gmail.com	\N	\N	\N	\N		\N	\N	3	5	2025-12-16 06:30:02.272598+03	\N	f	f	f	100	100	2	\N	t	f	16	refunded	completed	0	KSH	2025-12-16 06:31:05.655726+03	\N	2025-12-22 02:42:27.949658+03
1	2025-12-12 05:55:23.281217+03	2025-12-22 03:34:59.81613+03	\N	2	Kamau	simon	kamausimon217@gmail.com	\N	\N	\N	\N		\N	\N	6	10	2025-12-12 05:55:23.275307+03	\N	f	f	f	200	200	2	\N	t	f	32	refunded	completed	0	KSH	2025-12-16 06:17:10.7087+03	\N	2025-12-22 03:34:59.748264+03
4	2025-12-16 06:39:29.430907+03	2025-12-22 03:38:51.447978+03	\N	2	Kamau	simon	kamausimon217@gmail.com	\N	\N	\N	\N		\N	\N	6	10	2025-12-16 06:39:29.430841+03	\N	f	f	f	200	200	2	\N	t	f	32	refunded	completed	0	KSH	2025-12-16 06:40:30.573853+03	\N	2025-12-22 03:38:51.423148+03
5	2025-12-16 06:52:15.862149+03	2025-12-22 07:13:25.904221+03	\N	2	Kamau	simon	kamausimon217@gmail.com	\N	\N	\N	\N		\N	\N	6	10	2025-12-16 06:52:15.862081+03	\N	f	t	f	200	200	2	\N	t	f	32	cancelled	completed	0	KSH	2025-12-16 06:52:49.516706+03	2025-12-22 07:13:25.904091+03	2025-12-22 03:39:28.038964+03
\.


--
-- Data for Name: organizers; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.organizers (id, created_at, updated_at, deleted_at, account_id, name, about, email, phone, confirmation_key, facebook, twitter, logo_path, is_email_confirmed, show_twitter_widget, show_facebook_widget, tax_name, tax_value, tax_pin, charge_tax, page_header_bg_color, page_bg_color, page_text_color, enable_organizer_page, payment_gateway_id, bank_account_name, bank_account_number, bank_code, bank_country, is_payment_configured, is_verified, verification_status, rejection_reason, kyc_status, kyc_notes, kyc_completed_at) FROM stdin;
1	2025-12-09 06:01:55.261133+03	2025-12-10 00:36:51.251555+03	\N	3	Topstone	we host kikuyu love songs events	topstonewriters@gmail.com	0799500565				/uploads/logos/placeholder_3MEa33B__400x400.jpg	t	f	f	simon kamau	0	A01HHAFAVA	0				f	\N	Simon kamau	ai2Tn84uuJCPhS8SpY/Ef7H9TF8FuyU08WsUON4NjV6vachoW2zKjN4=	0350Um5R8X8MiPLw8PIjAkGNKVwcc/z+1YQOdOMOU5a3DguV	kenya	f	t	approved		completed	The meeting with the organizer has been scheduled\n\n---\n\nThe meeting with the organizer has been scheduled\n\n---\n\nThe meeting with the organizer has been completed	2025-12-09T21:58:01+03:00
\.


--
-- Data for Name: password_reset_attempts; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.password_reset_attempts (id, created_at, updated_at, deleted_at, password_reset_id, ip_address, user_agent, attempted_at, was_successful, token_valid, not_expired, ip_matched, rate_limit_passed, failure_reason, error_code, country, city, isp, response_time_ms) FROM stdin;
1	2025-12-08 06:12:55.301749+03	2025-12-08 06:12:55.301749+03	\N	1	[::1]	PostmanRuntime/7.49.1	2025-12-08 06:12:55.254722+03	f	f	f	t	t	Token expired	TOKEN_EXPIRED	\N	\N	\N	\N
2	2025-12-08 06:14:22.570444+03	2025-12-08 06:14:22.570444+03	\N	1	[::1]	PostmanRuntime/7.49.1	2025-12-08 06:14:22.540961+03	f	f	f	t	t	Token expired	TOKEN_EXPIRED	\N	\N	\N	\N
3	2025-12-08 06:16:09.903872+03	2025-12-08 06:16:09.903872+03	\N	2	[::1]	PostmanRuntime/7.49.1	2025-12-08 06:16:09.28141+03	t	t	t	t	t	\N	\N	\N	\N	\N	\N
4	2025-12-08 06:17:37.27685+03	2025-12-08 06:17:37.27685+03	\N	3	[::1]	PostmanRuntime/7.49.1	2025-12-08 06:17:36.59871+03	t	t	t	t	t	\N	\N	\N	\N	\N	\N
\.


--
-- Data for Name: password_resets; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.password_resets (id, created_at, updated_at, deleted_at, token, email, status, method, user_id, account_id, ip_address, user_agent, attempt_count, max_attempts, expires_at, issued_at, used_at, revoked_at, last_attempt_at, original_ip, used_from_ip, same_ip_required, require_current_password, require_two_factor, is_security_reset, requested_by, approved_by, rate_limit_key, previous_reset_at, cooldown_until, reset_reason, admin_notes, user_message, should_cleanup, cleanup_after) FROM stdin;
1	2025-12-08 05:56:57.905314+03	2025-12-08 06:14:22.668669+03	\N	51e983109f3427ded0ac00589974acc7	kamausimon217@gmail.com	expired	email	2	2	[::1]	PostmanRuntime/7.49.1	0	3	2025-12-08 06:11:57.877944+03	2025-12-08 05:56:57.877944+03	\N	\N	\N	[::1]	\N	f	f	f	f	\N	\N		\N	\N	\N	\N	\N	t	2025-12-15 05:56:57.877944+03
2	2025-12-08 06:15:37.296117+03	2025-12-08 06:16:09.823267+03	\N	e1aa28bf3b2ce0da989d2057d07aadb4	kamausimon217@gmail.com	used	email	2	2	[::1]	PostmanRuntime/7.49.1	0	3	2025-12-08 06:30:36.028416+03	2025-12-08 06:15:36.028416+03	2025-12-08 06:16:09.767182+03	\N	\N	[::1]	[::1]	f	f	f	f	\N	\N		\N	\N	\N	\N	\N	t	2025-12-15 06:15:36.028416+03
3	2025-12-08 06:17:04.220349+03	2025-12-08 06:17:37.234576+03	\N	dc948e2cf0e06967ebb59e42b0c0aa7c	kamausimon217@gmail.com	used	email	2	2	[::1]	PostmanRuntime/7.49.1	0	3	2025-12-08 06:32:04.204297+03	2025-12-08 06:17:04.204297+03	2025-12-08 06:17:37.154638+03	\N	\N	[::1]	[::1]	f	f	f	f	\N	\N		\N	\N	\N	\N	\N	t	2025-12-15 06:17:04.204297+03
\.


--
-- Data for Name: payment_gateways; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.payment_gateways (id, created_at, updated_at, deleted_at, provider_name, provider_url, is_on_site, can_refund, name) FROM stdin;
\.


--
-- Data for Name: payment_methods; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.payment_methods (id, created_at, updated_at, deleted_at, account_id, type, status, display_name, nickname, is_default, card_brand, card_last4, card_expiry_month, card_expiry_year, card_country, card_fingerprint, mpesa_phone_number, mpesa_account_name, bank_account_last4, bank_name, bank_code, bank_account_holder, stripe_payment_method_id, stripe_customer_id, external_payment_method_id, is_verified, verified_at, last_used_at, failure_count, last_failure_at, billing_address, metadata) FROM stdin;
\.


--
-- Data for Name: payment_records; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.payment_records (id, created_at, updated_at, deleted_at, amount, currency, type, status, order_id, event_id, account_id, organizer_id, payment_gateway_id, external_transaction_id, external_reference, gateway_response_code, initiated_at, processed_at, completed_at, failed_at, description, notes, ip_address, user_agent, platform_fee_amount, gateway_fee_amount, net_amount, parent_record_id, reconciled_at, reconciliation_ref) FROM stdin;
1	2025-12-15 06:31:13.604022+03	2025-12-15 06:31:13.604022+03	\N	248	KES	customer_payment	pending	1	2	2	\N	\N	\N	ORD-1-1765769473	\N	2025-12-15 06:31:13.587776+03	\N	\N	\N	Payment for Order #1	\N	\N	\N	0	0	0	\N	\N	\N
2	2025-12-15 06:40:33.301213+03	2025-12-15 06:40:33.301213+03	\N	248	KES	customer_payment	pending	1	2	2	\N	\N	\N	ORD-1-1765770033	\N	2025-12-15 06:40:33.279306+03	\N	\N	\N	Payment for Order #1	\N	\N	\N	0	0	0	\N	\N	\N
3	2025-12-15 06:44:46.311378+03	2025-12-15 06:44:46.311378+03	\N	248	KES	customer_payment	pending	1	2	2	\N	\N	\N	ORD-1-1765770286	\N	2025-12-15 06:44:46.281106+03	\N	\N	\N	Payment for Order #1	\N	\N	\N	0	0	0	\N	\N	\N
4	2025-12-15 06:48:44.459725+03	2025-12-15 06:48:44.459725+03	\N	248	KES	customer_payment	pending	1	2	2	\N	\N	\N	ORD-1-1765770524	\N	2025-12-15 06:48:44.446394+03	\N	\N	\N	Payment for Order #1	\N	\N	\N	0	0	0	\N	\N	\N
5	2025-12-15 06:51:21.996916+03	2025-12-15 06:51:21.996916+03	\N	248	KES	customer_payment	pending	1	2	2	\N	\N	\N	ORD-1-1765770681	\N	2025-12-15 06:51:21.972567+03	\N	\N	\N	Payment for Order #1	\N	\N	\N	0	0	0	\N	\N	\N
6	2025-12-15 06:56:04.511713+03	2025-12-15 06:56:04.511713+03	\N	248	KES	customer_payment	pending	1	2	2	\N	\N	\N	ORD-1-1765770964	\N	2025-12-15 06:56:04.477344+03	\N	\N	\N	Payment for Order #1	\N	\N	\N	0	0	0	\N	\N	\N
7	2025-12-15 07:03:03.320561+03	2025-12-15 07:03:03.320561+03	\N	10	KES	customer_payment	pending	1	2	2	\N	\N	\N	ORD-1-1765771383	\N	2025-12-15 07:03:03.306071+03	\N	\N	\N	Payment for Order #1	\N	\N	\N	0	0	0	\N	\N	\N
8	2025-12-15 07:03:43.079461+03	2025-12-15 07:03:43.079461+03	\N	1000	KES	customer_payment	pending	1	2	2	\N	\N	\N	ORD-1-1765771423	\N	2025-12-15 07:03:43.048298+03	\N	\N	\N	Payment for Order #1	\N	\N	\N	0	0	0	\N	\N	\N
9	2025-12-15 07:04:34.334251+03	2025-12-15 07:04:34.334251+03	\N	1000	KES	customer_payment	pending	1	2	2	\N	\N	\N	ORD-1-1765771474	\N	2025-12-15 07:04:34.315305+03	\N	\N	\N	Payment for Order #1	\N	\N	\N	0	0	0	\N	\N	\N
10	2025-12-16 04:24:17.894219+03	2025-12-16 04:24:17.894219+03	\N	248	KES	customer_payment	pending	1	2	2	\N	\N	\N	ORD-1-1765848257	\N	2025-12-16 04:24:17.85424+03	\N	\N	\N	Payment for Order #1	\N	\N	\N	0	0	0	\N	\N	\N
11	2025-12-16 05:07:25.265727+03	2025-12-16 05:07:25.265727+03	\N	248	KES	customer_payment	pending	1	2	2	\N	\N	\N	ORD-1-1765850845	\N	2025-12-16 05:07:25.239687+03	\N	\N	\N	Payment for Order #1	\N	\N	\N	0	0	0	\N	\N	\N
12	2025-12-16 05:09:53.582924+03	2025-12-16 05:09:53.582924+03	\N	248	KES	customer_payment	pending	1	2	2	\N	\N	\N	ORD-1-1765850993	\N	2025-12-16 05:09:53.527586+03	\N	\N	\N	Payment for Order #1	\N	\N	\N	0	0	0	\N	\N	\N
13	2025-12-16 05:11:16.259456+03	2025-12-16 05:11:16.259456+03	\N	248	KES	customer_payment	pending	1	2	2	\N	\N	\N	ORD-1-1765851076	\N	2025-12-16 05:11:16.240229+03	\N	\N	\N	Payment for Order #1	\N	\N	\N	0	0	0	\N	\N	\N
14	2025-12-16 05:14:47.324307+03	2025-12-16 05:14:47.324307+03	\N	248	KES	customer_payment	pending	1	2	2	\N	\N	\N	ORD-1-1765851287	\N	2025-12-16 05:14:47.300542+03	\N	\N	\N	Payment for Order #1	\N	\N	\N	0	0	0	\N	\N	\N
15	2025-12-16 05:19:26.397172+03	2025-12-16 05:19:26.397172+03	\N	248	KES	customer_payment	pending	1	2	2	\N	\N	\N	ORD-1-1765851566	\N	2025-12-16 05:19:26.372395+03	\N	\N	\N	Payment for Order #1	\N	\N	\N	0	0	0	\N	\N	\N
16	2025-12-16 05:35:37.014572+03	2025-12-16 05:35:37.014572+03	\N	248	KES	customer_payment	pending	1	2	2	\N	\N	\N	ORD-1-1765852536	\N	2025-12-16 05:35:36.99645+03	\N	\N	\N	Payment for Order #1	\N	\N	\N	0	0	0	\N	\N	\N
17	2025-12-16 05:39:56.600625+03	2025-12-16 05:39:56.600625+03	\N	248	KES	customer_payment	pending	1	2	2	\N	\N	\N	ORD-1-1765852796	\N	2025-12-16 05:39:56.54775+03	\N	\N	\N	Payment for Order #1	\N	\N	\N	0	0	0	\N	\N	\N
18	2025-12-16 05:43:18.035406+03	2025-12-16 05:43:18.035406+03	\N	248	KES	customer_payment	pending	1	2	2	\N	\N	\N	ORD-1-1765852997	\N	2025-12-16 05:43:17.997158+03	\N	\N	\N	Payment for Order #1	\N	\N	\N	0	0	0	\N	\N	\N
19	2025-12-16 05:59:46.384538+03	2025-12-16 05:59:46.384538+03	\N	248	KES	customer_payment	pending	1	2	2	\N	\N	\N	ORD-1-1765853986	\N	2025-12-16 05:59:46.356133+03	\N	\N	\N	Payment for Order #1	\N	\N	\N	0	0	0	\N	\N	\N
20	2025-12-16 06:16:43.28151+03	2025-12-16 06:17:10.708816+03	\N	248	KES	customer_payment	completed	1	2	2	\N	\N		ORD-1-1765855003	\N	2025-12-16 06:16:43.235064+03	2025-12-16 06:17:00.611265+03	2025-12-16 06:17:10.7087+03	\N	Payment for Order #1	\N	\N	\N	0	9	239	\N	\N	\N
21	2025-12-16 06:21:37.985489+03	2025-12-16 06:22:06.875315+03	\N	248	KES	customer_payment	completed	1	2	2	\N	\N		ORD-1-1765855297	\N	2025-12-16 06:21:37.971257+03	2025-12-16 06:21:58.506569+03	2025-12-16 06:22:06.875228+03	\N	Payment for Order #1	\N	\N	\N	0	9	239	\N	\N	\N
22	2025-12-16 06:24:50.548814+03	2025-12-16 06:25:11.70845+03	\N	248	KES	customer_payment	completed	1	2	2	\N	\N		ORD-1-1765855490	\N	2025-12-16 06:24:50.525996+03	2025-12-16 06:25:03.574152+03	2025-12-16 06:25:11.708367+03	\N	Payment for Order #1	\N	\N	\N	0	9	239	\N	\N	\N
23	2025-12-16 06:30:43.464842+03	2025-12-16 06:31:05.655821+03	\N	119	KES	customer_payment	completed	3	2	2	\N	\N		ORD-3-1765855843	\N	2025-12-16 06:30:43.438693+03	2025-12-16 06:30:57.508661+03	2025-12-16 06:31:05.655726+03	\N	Payment for Order #3	\N	\N	\N	0	5	112	\N	\N	\N
24	2025-12-16 06:40:00.404316+03	2025-12-16 06:40:30.573926+03	\N	238	KES	customer_payment	completed	4	2	2	\N	\N		ORD-4-1765856400	\N	2025-12-16 06:40:00.387488+03	2025-12-16 06:40:21.596864+03	2025-12-16 06:40:30.573853+03	\N	Payment for Order #4	\N	\N	\N	0	9	229	\N	\N	\N
25	2025-12-16 06:52:27.09471+03	2025-12-16 06:52:49.51679+03	\N	238	KES	customer_payment	completed	5	2	2	\N	\N		ORD-5-1765857147	\N	2025-12-16 06:52:27.073626+03	2025-12-16 06:52:40.779912+03	2025-12-16 06:52:49.516706+03	\N	Payment for Order #5	\N	\N	\N	0	9	229	\N	\N	\N
26	2025-12-16 06:55:27.71153+03	2025-12-16 06:55:47.655675+03	\N	238	KES	customer_payment	completed	6	2	2	\N	\N		ORD-6-1765857327	\N	2025-12-16 06:55:27.697086+03	2025-12-16 06:55:39.722508+03	2025-12-16 06:55:47.655589+03	\N	Payment for Order #6	\N	\N	\N	0	9	229	\N	\N	\N
27	2025-12-16 07:08:31.601845+03	2025-12-16 07:08:39.724376+03	\N	238	KES	customer_payment	completed	7	2	2	\N	\N		ORD-7-1765858111	\N	2025-12-16 07:08:31.576099+03	2025-12-16 07:08:31.565118+03	2025-12-16 07:08:39.724294+03	\N	Payment for Order #7	\N	\N	\N	0	9	229	\N	\N	\N
28	2025-12-16 07:14:25.071258+03	2025-12-16 07:14:29.699933+03	\N	119	KES	customer_payment	completed	7	2	2	\N	\N		ORD-7-1765858465	\N	2025-12-16 07:14:25.05194+03	2025-12-16 07:14:21.762153+03	2025-12-16 07:14:29.699861+03	\N	Payment for Order #7	\N	\N	\N	0	5	112	\N	\N	\N
29	2025-12-16 07:24:07.429913+03	2025-12-16 07:24:17.827086+03	\N	119	KES	customer_payment	completed	9	2	2	\N	\N		ORD-9-1765859047	\N	2025-12-16 07:24:07.411712+03	2025-12-16 07:24:24.211615+03	2025-12-16 07:24:17.826993+03	\N	Payment for Order #9	\N	\N	\N	0	5	112	\N	\N	\N
\.


--
-- Data for Name: payment_transactions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.payment_transactions (id, created_at, updated_at, deleted_at, amount, currency, type, status, order_id, payment_gateway_id, organizer_id, external_transaction_id, external_reference, processed_at, settled_at, description, notes, parent_transaction_id) FROM stdin;
1	2025-12-16 06:17:10.733927+03	2025-12-16 06:17:10.733927+03	\N	248	KES	payment	completed	1	\N	\N		ORD-1-1765855003	2025-12-16 06:17:10.7087+03	\N	Payment via CARD-PAYMENT	\N	\N
2	2025-12-16 06:22:06.910831+03	2025-12-16 06:22:06.910831+03	\N	248	KES	payment	completed	1	\N	\N		ORD-1-1765855297	2025-12-16 06:22:06.875228+03	\N	Payment via CARD-PAYMENT	\N	\N
3	2025-12-16 06:25:11.712672+03	2025-12-16 06:25:11.712672+03	\N	248	KES	payment	completed	1	\N	\N		ORD-1-1765855490	2025-12-16 06:25:11.708367+03	\N	Payment via CARD-PAYMENT	\N	\N
4	2025-12-16 06:31:05.668982+03	2025-12-16 06:31:05.668982+03	\N	119	KES	payment	completed	3	\N	\N		ORD-3-1765855843	2025-12-16 06:31:05.655726+03	\N	Payment via CARD-PAYMENT	\N	\N
5	2025-12-16 06:40:30.577345+03	2025-12-16 06:40:30.577345+03	\N	238	KES	payment	completed	4	\N	\N		ORD-4-1765856400	2025-12-16 06:40:30.573853+03	\N	Payment via CARD-PAYMENT	\N	\N
6	2025-12-16 06:52:49.520596+03	2025-12-16 06:52:49.520596+03	\N	238	KES	payment	completed	5	\N	\N		ORD-5-1765857147	2025-12-16 06:52:49.516706+03	\N	Payment via CARD-PAYMENT	\N	\N
7	2025-12-16 06:55:47.659251+03	2025-12-16 06:55:47.659251+03	\N	238	KES	payment	completed	6	\N	\N		ORD-6-1765857327	2025-12-16 06:55:47.655589+03	\N	Payment via CARD-PAYMENT	\N	\N
8	2025-12-16 07:08:39.728079+03	2025-12-16 07:08:39.728079+03	\N	238	KES	payment	completed	7	\N	\N		ORD-7-1765858111	2025-12-16 07:08:39.724294+03	\N	Payment via CARD-PAYMENT	\N	\N
9	2025-12-16 07:14:29.709951+03	2025-12-16 07:14:29.709951+03	\N	119	KES	payment	completed	7	\N	\N		ORD-7-1765858465	2025-12-16 07:14:29.699861+03	\N	Payment via CARD-PAYMENT	\N	\N
10	2025-12-16 07:24:17.833566+03	2025-12-16 07:24:17.833566+03	\N	119	KES	payment	completed	9	\N	\N		ORD-9-1765859047	2025-12-16 07:24:17.826993+03	\N	Payment via CARD-PAYMENT	\N	\N
\.


--
-- Data for Name: payout_accounts; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.payout_accounts (id, created_at, updated_at, deleted_at, organizer_id, account_type, status, display_name, is_default, bank_name, bank_code, bank_branch, bank_country, account_number, account_holder_name, mobile_provider, mobile_phone_number, mobile_account_name, paypal_email, stripe_account_id, stripe_country, currency, is_verified, verified_at, verified_by, verification_notes, document_paths, verification_token, total_payouts_count, total_payouts_amount, last_payout_at, last_payout_amount, failed_payouts_count, last_failure_at, last_failure_reason, requires_kyc, kyc_status, kyc_completed_at, is_suspicious_activity, suspicion_reason, reviewed_by, reviewed_at, daily_payout_limit, monthly_payout_limit, requires_approval, external_account_id, external_metadata, address_line1, address_line2, city, state, postal_code, country, notes) FROM stdin;
\.


--
-- Data for Name: promotion_rules; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.promotion_rules (id, promotion_id, rule_type, rule_operator, rule_value, error_message, is_active, execution_order, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: promotion_usages; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.promotion_usages (id, created_at, updated_at, deleted_at, promotion_id, order_id, account_id, discount_amount, original_amount, final_amount, used_at, ip_address, user_agent, validation_time, cache_hit) FROM stdin;
\.


--
-- Data for Name: promotions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.promotions (id, created_at, updated_at, deleted_at, code, name, description, type, status, target, discount_percentage, discount_amount, free_quantity, minimum_purchase, maximum_discount, event_id, ticket_class_ids, event_categories, organizer_id, start_date, end_date, early_bird_cutoff, usage_limit, usage_count, per_user_limit, per_order_limit, is_unlimited, precomputed_active, last_usage_check, first_time_customers, minimum_age, allowed_user_ids, excluded_user_ids, created_by, is_public, requires_approval, total_revenue, total_discount, conversion_rate, internal_notes, marketing_tags) FROM stdin;
1	2025-12-24 03:55:17.268593+03	2025-12-24 04:02:29.075756+03	2025-12-24 04:03:37.906733+03	HZ4YCEQF	kikuyu love songs sale	kikuyu love sale flash sale	percentage	draft	specific_ticket	50	\N	\N	\N	\N	\N			1	2025-12-24 03:00:00+03	2025-12-30 03:00:00+03	\N	200	0	2	1	f	f	\N	f	\N			3	t	f	0	0	\N	\N	
3	2025-12-24 04:06:27.033347+03	2025-12-24 04:21:48.562357+03	\N	NIVLBJ5Z	kikuyu love songs sale (Copy)	50% off on all tickets	percentage	cancelled	specific_ticket	50	\N	\N	\N	\N	\N			1	2025-12-24 04:06:27.003719+03	2026-01-24 04:06:27.003719+03	\N	200	0	2	1	f	f	\N	f	\N			3	t	f	0	0	\N	\N	
4	2025-12-24 11:03:22.401259+03	2025-12-24 11:03:22.401259+03	2025-12-24 11:20:37.390914+03	4MRSCKID	kikuyu love songs sale	30% off on all tickets	percentage	draft	specific_ticket	30	\N	\N	\N	\N	\N	[1,2,3]		1	2025-12-24 03:00:00+03	2025-12-30 03:00:00+03	\N	200	0	2	1	f	f	\N	f	\N			3	t	f	0	0	\N	\N	
5	2025-12-24 11:12:01.234255+03	2025-12-24 11:12:01.234255+03	2025-12-24 11:20:57.012215+03	CB62435V	kikuyu love songs sale	30% off on all tickets	percentage	draft	specific_ticket	30	\N	\N	\N	\N	\N	[1,2,3]		1	2025-12-24 03:00:00+03	2025-12-30 03:00:00+03	\N	200	0	2	1	f	f	\N	f	\N			3	t	f	0	0	\N	\N	
6	2025-12-24 11:21:15.103959+03	2025-12-24 11:26:03.908388+03	\N	KY5CPBBJ	kikuyu love songs sale	30% off on all tickets	percentage	active	specific_ticket	30	\N	\N	\N	\N	2	[1,2,3]		1	2025-12-24 03:00:00+03	2025-12-30 03:00:00+03	\N	200	0	2	1	f	t	\N	f	\N			3	t	f	0	0	\N	\N	
2	2025-12-24 04:03:58.109727+03	2025-12-24 11:26:38.355575+03	\N	LT27N4NN	kikuyu love songs sale	50% off on all tickets	percentage	paused	specific_ticket	50	\N	\N	\N	\N	\N			1	2025-12-24 03:00:00+03	2025-12-31 04:00:00+03	\N	200	0	2	1	f	f	\N	f	\N			3	t	f	0	0	\N	\N	
\.


--
-- Data for Name: recovery_codes; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.recovery_codes (id, created_at, updated_at, deleted_at, two_factor_auth_id, code_hash, used, used_at, used_from_ip) FROM stdin;
\.


--
-- Data for Name: refund_line_items; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.refund_line_items (id, created_at, updated_at, deleted_at, refund_record_id, order_item_id, ticket_id, quantity, refund_amount, reason, description) FROM stdin;
\.


--
-- Data for Name: refund_records; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.refund_records (id, created_at, updated_at, deleted_at, refund_number, refund_type, refund_reason, status, order_id, event_id, account_id, organizer_id, original_amount, refund_amount, organizer_impact, currency, payment_gateway_id, external_refund_id, requested_by, approved_by, requested_at, approved_at, processed_at, completed_at, failed_at, affects_settlement, settlement_adjusted, description, internal_notes, rejection_reason) FROM stdin;
\.


--
-- Data for Name: reserved_tickets; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.reserved_tickets (id, created_at, updated_at, deleted_at, ticket_id, event_id, quantity_reserved, expires, session_id) FROM stdin;
1	2025-12-28 11:58:32.281234+03	2025-12-28 12:13:31.298532+03	2025-12-28 18:20:14.446489+03	1	2	2	2025-12-28 12:23:32.281182+03	user_2
2	2025-12-28 12:04:24.567319+03	2025-12-28 12:04:24.567319+03	2025-12-28 18:20:14.446489+03	2	2	2	2025-12-28 12:19:24.567255+03	user_2
3	2025-12-28 18:23:42.698023+03	2025-12-28 18:23:42.698023+03	2025-12-28 18:39:17.81095+03	2	2	2	2025-12-28 18:38:42.697751+03	user_2
4	2025-12-28 18:59:02.804719+03	2025-12-28 18:59:02.804719+03	2025-12-28 18:59:34.422926+03	1	2	1	2025-12-28 19:14:02.804629+03	user_2
\.


--
-- Data for Name: reset_configurations; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.reset_configurations (id, created_at, updated_at, deleted_at, token_length, token_expiry_minutes, token_algorithm, max_attempts_per_token, max_requests_per_hour, max_requests_per_ip, cooldown_minutes, require_same_ip, allow_vp_ns, block_known_proxies, cleanup_after_days, keep_audit_days, auto_cleanup_enabled, send_confirmation_email, notify_on_suspicious, log_all_attempts, email_reset_enabled, sms_reset_enabled, admin_reset_enabled, config_name, description, is_active, created_by, last_modified_by) FROM stdin;
2	2025-12-08 06:02:33.103725+03	2025-12-08 06:02:33.103725+03	\N	32	15	random	3	5	10	30	f	t	f	7	90	t	t	t	t	t	f	t	default	Default password reset configuration	t	1	\N
\.


--
-- Data for Name: security_metrics; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.security_metrics (id, created_at, updated_at, deleted_at, event_type, severity, "timestamp", date, hour, ip_address, user_agent, country, account_id, user_id, description, raw_data, risk_score, is_blocked, action_taken, is_resolved, resolved_at, resolved_by, resolution) FROM stdin;
\.


--
-- Data for Name: settlement_items; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.settlement_items (id, created_at, updated_at, deleted_at, settlement_record_id, organizer_id, event_id, event_status, event_end_date, event_verified_at, has_disputes, refund_amount_issued, chargeback_amount, risk_hold_applied, risk_hold_reason, gross_amount, platform_fee_amount, refund_deduction, adjustment_amount, net_amount, currency, status, external_transaction_id, external_reference, bank_account_number, bank_name, bank_code, account_holder_name, processed_at, completed_at, failed_at, failure_reason, description, notes) FROM stdin;
\.


--
-- Data for Name: settlement_payment_records; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.settlement_payment_records (settlement_item_id, payment_record_id) FROM stdin;
\.


--
-- Data for Name: settlement_records; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.settlement_records (id, created_at, updated_at, deleted_at, settlement_batch_id, description, status, frequency, trigger, event_id, event_completed_at, event_completion_verified, holding_period_days, holding_period_start_date, holding_period_end_date, earliest_payout_date, has_active_disputes, dispute_count, chargeback_count, refund_amount, withholding_reason, period_start_date, period_end_date, total_organizers, total_amount, total_payment_records, currency, initiated_by, approved_by, approved_at, processed_at, completed_at, failed_at, external_batch_id, payment_gateway_id, notes, internal_reference, risk_score) FROM stdin;
\.


--
-- Data for Name: support_ticket_comments; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.support_ticket_comments (id, created_at, updated_at, deleted_at, ticket_id, user_id, comment, is_internal, author_name, author_email) FROM stdin;
1	2026-01-08 12:28:47.159108+03	2026-01-08 12:28:47.159108+03	\N	1	2	Hello, I am facing an issue with my account. Can you please help me resolve it?	f	Kamau Simon	kamausimon217@gmail.com
2	2026-01-08 13:46:28.013679+03	2026-01-08 13:46:28.013679+03	\N	1	2	Hello, I am facing an issue with my account. Can you please help me resolve it?	f	Kamau Simon	kamausimon217@gmail.com
\.


--
-- Data for Name: support_tickets; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.support_tickets (id, created_at, updated_at, deleted_at, ticket_number, subject, description, category, priority, status, user_id, email, name, phone_number, order_id, event_id, organizer_id, assigned_to_id, resolved_at, resolved_by_id, resolution_notes, ai_classified, a_ipriority, ai_confidence_score, ai_reasoning) FROM stdin;
1	2026-01-08 12:09:10.485068+03	2026-01-08 12:55:13.284568+03	\N	TKT-20260108-5685	payment not going thru	I am having trouble with my payment not going through. I have tried multiple times and it keeps failing. Can you please help me resolve this issue?	payment	medium	resolved	2	kamausimon217@gmail.com	Kamau siMon		\N	\N	\N	\N	2026-01-08 12:55:13.266704+03	4	Ticket resolved, go ahead and retry making the payment	f		0.0000	
2	2026-01-08 13:32:06.987952+03	2026-01-08 13:32:06.987952+03	\N	TKT-20260108-7316	event issue	I am having trouble with my event not going thru...	event	medium	open	4	kamausimon217@gmail.com	Kamau siMon		\N	\N	\N	\N	\N	\N		f		0.0000	
3	2026-01-08 13:44:38.556964+03	2026-01-08 13:47:46.460609+03	\N	TKT-20260108-6762	event issue	I am having trouble with my event not going thru...	event	medium	closed	2	kamausimon217@gmail.com	Kamau siMon		\N	\N	\N	\N	2026-01-08 13:47:46.423884+03	4	Ticket resolved, go ahead and retry making the payment	f		0.0000	
\.


--
-- Data for Name: system_metrics; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.system_metrics (id, created_at, updated_at, deleted_at, metric_name, metric_type, granularity, "timestamp", date, hour, day_of_week, week, month, year, value, count, sum, min, max, event_id, organizer_id, account_id, country, region, city, dimensions, tags, source, version, is_estimate) FROM stdin;
\.


--
-- Data for Name: ticket_classes; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.ticket_classes (id, created_at, updated_at, deleted_at, event_id, name, description, price, currency, max_per_order, min_per_order, quantity_available, quantity_sold, version, start_sale_date, end_sale_date, sales_volume, organizer_fees_volume, is_paused, is_hidden, sort_order, requires_approval) FROM stdin;
3	2025-12-12 02:19:23.094756+03	2025-12-12 02:19:23.094756+03	2025-12-12 02:21:01.604103+03	2	vvip	vvip price ticket	250	KSH	5	1	500	0	0	2025-12-17 03:00:00+03	2025-12-29 03:00:00+03	0	0	f	f	0	f
4	2025-12-12 02:21:23.297191+03	2025-12-12 05:27:33.480702+03	\N	2	vvip	vvip price ticket	300	KSH	5	1	500	0	0	2025-12-17 03:00:00+03	2025-12-29 03:00:00+03	0	0	f	f	0	f
2	2025-12-12 02:07:24.385986+03	2025-12-12 06:08:45.558602+03	\N	2	vip	vip price ticket	200	KSH	5	1	2000	1	1	2025-12-17 03:00:00+03	2025-12-29 03:00:00+03	0	0	f	f	0	f
1	2025-12-12 02:03:42.273959+03	2025-12-22 07:13:25.955937+03	\N	2	regular	regular price ticket	100	KSH	5	1	2000	10	8	2025-12-11 08:30:00+03	2025-12-29 03:00:00+03	0	0	f	f	0	f
5	2026-01-10 22:29:08.887574+03	2026-01-10 22:29:08.887574+03	\N	4	vvip	vvip price ticket	10000	KSH	5	1	200	0	0	2026-01-11 03:00:00+03	2026-12-29 03:00:00+03	0	0	f	f	0	f
\.


--
-- Data for Name: ticket_orders; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.ticket_orders (id, created_at, updated_at, deleted_at, order_id, ticket_id) FROM stdin;
\.


--
-- Data for Name: ticket_transfer_histories; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.ticket_transfer_histories (id, created_at, updated_at, deleted_at, ticket_id, from_holder_name, from_holder_email, to_holder_name, to_holder_email, transferred_by, transferred_at, transfer_reason, ip_address, user_agent) FROM stdin;
1	2025-12-23 14:33:23.870315+03	2025-12-23 14:33:23.870315+03	\N	1	Kamau simon	kamausimon217@gmail.com	Admin Harsha	harshasg10@gmail.com	2	2025-12-23 14:33:23.842272+03		[::1]:34296	PostmanRuntime/7.51.0
2	2025-12-23 14:39:17.066762+03	2025-12-23 14:39:17.066762+03	\N	2	Kamau simon	kamausimon217@gmail.com	Admin Harsha	harshasg10@gmail.com	2	2025-12-23 14:39:17.053313+03		[::1]:48340	PostmanRuntime/7.51.0
\.


--
-- Data for Name: tickets; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.tickets (id, created_at, updated_at, deleted_at, order_item_id, ticket_number, qr_code, barcode_data, holder_name, holder_email, status, checked_in_at, checked_in_by, used_at, refunded_at, pdf_path) FROM stdin;
4	2025-12-16 06:40:30.650045+03	2025-12-16 06:40:30.650045+03	\N	4	TKT-2-4-4-0-1765856430	TICKET:EVENT2:ORDER4:IDX0:TIME1765856430		Kamau simon	kamausimon217@gmail.com	active	\N	\N	\N	\N	\N
5	2025-12-16 06:40:30.653519+03	2025-12-16 06:40:30.653519+03	\N	4	TKT-2-4-4-1-1765856430	TICKET:EVENT2:ORDER4:IDX1:TIME1765856430		Kamau simon	kamausimon217@gmail.com	active	\N	\N	\N	\N	\N
8	2025-12-16 06:55:47.69036+03	2025-12-16 06:55:47.69036+03	\N	6	TKT-2-6-6-0-1765857347	TICKET:EVENT2:ORDER6:IDX0:TIME1765857347		Kamau simon	kamausimon217@gmail.com	active	\N	\N	\N	\N	\N
9	2025-12-16 06:55:47.692391+03	2025-12-16 06:55:47.692391+03	\N	6	TKT-2-6-6-1-1765857347	TICKET:EVENT2:ORDER6:IDX1:TIME1765857347		Kamau simon	kamausimon217@gmail.com	active	\N	\N	\N	\N	\N
10	2025-12-16 07:08:39.777025+03	2025-12-16 07:08:40.13987+03	\N	7	TKT-2-7-7-0-1765858119	TICKET:EVENT2:ORDER7:IDX0:TIME1765858119		Kamau simon	kamausimon217@gmail.com	active	\N	\N	\N	\N	storage/tickets/7/ticket_TKT-2-7-7-0-1765858119.pdf
6	2025-12-16 06:52:49.560413+03	2025-12-22 07:13:25.910143+03	\N	5	TKT-2-5-5-0-1765857169	TICKET:EVENT2:ORDER5:IDX0:TIME1765857169		Kamau simon	kamausimon217@gmail.com	cancelled	\N	\N	\N	\N	./tickets/pdfs/ticket_TKT-2-5-5-0-1765857169.pdf
7	2025-12-16 06:52:49.562411+03	2025-12-22 07:13:25.910143+03	\N	5	TKT-2-5-5-1-1765857169	TICKET:EVENT2:ORDER5:IDX1:TIME1765857169		Kamau simon	kamausimon217@gmail.com	cancelled	\N	\N	\N	\N	./tickets/pdfs/ticket_TKT-2-5-5-1-1765857169.pdf
12	2025-12-16 07:24:17.978668+03	2025-12-23 15:42:23.946204+03	\N	9	TKT-2-9-9-0-1765859057	TICKET:EVENT2:ORDER9:IDX0:TIME1765859057		Kamau simon	kamausimon217@gmail.com	used	2025-12-23 15:42:23.924624+03	3	2025-12-23 15:42:23.924624+03	\N	storage/tickets/9/ticket_TKT-2-9-9-0-1765859057.pdf
1	2025-12-16 06:17:10.85702+03	2025-12-23 15:55:31.441527+03	\N	1	TKT-2-1-1-0-1765855030	TICKET:EVENT2:ORDER1:IDX0:TIME1765855030		Admin Harsha	harshasg10@gmail.com	used	2025-12-23 15:55:31.420893+03	3	2025-12-23 15:55:31.420893+03	\N	\N
2	2025-12-16 06:17:10.956048+03	2025-12-24 02:05:54.080079+03	\N	1	TKT-2-1-1-1-1765855030	TICKET:EVENT2:ORDER1:IDX1:TIME1765855030		Admin Harsha	harshasg10@gmail.com	used	\N	\N	\N	\N	\N
11	2025-12-16 07:08:39.861365+03	2025-12-24 02:05:54.080079+03	\N	7	TKT-2-7-7-1-1765858119	TICKET:EVENT2:ORDER7:IDX1:TIME1765858119		Kamau simon	kamausimon217@gmail.com	used	\N	\N	\N	\N	storage/tickets/7/ticket_TKT-2-7-7-1-1765858119.pdf
3	2025-12-16 06:31:05.723819+03	2025-12-24 12:03:14.834652+03	\N	3	TKT-2-3-3-0-1765855865	TICKET:EVENT2:ORDER3:IDX0:TIME1765855865		Kamau simon	kamausimon217@gmail.com	used	2025-12-24 12:03:14.766481+03	3	2025-12-24 12:03:14.766481+03	\N	\N
\.


--
-- Data for Name: timezones; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.timezones (id, created_at, updated_at, deleted_at, name, display_name, "offset", iana_name, is_active) FROM stdin;
1	2025-12-11 02:41:00.078267+03	2025-12-11 02:41:00.078267+03	\N	UTC	UTC - Coordinated Universal Time	+00:00	UTC	t
2	2025-12-11 02:41:00.144107+03	2025-12-11 02:41:00.144107+03	\N	EAT	East Africa Time (Nairobi, Kampala, Dar es Salaam)	+03:00	Africa/Nairobi	t
3	2025-12-11 02:41:00.201045+03	2025-12-11 02:41:00.201045+03	\N	WAT	West Africa Time (Lagos, Accra)	+01:00	Africa/Lagos	t
4	2025-12-11 02:41:00.27317+03	2025-12-11 02:41:00.27317+03	\N	CAT	Central Africa Time (Johannesburg)	+02:00	Africa/Johannesburg	t
5	2025-12-11 02:41:00.321299+03	2025-12-11 02:41:00.321299+03	\N	EST	Eastern Standard Time (New York)	-05:00	America/New_York	t
6	2025-12-11 02:41:00.385839+03	2025-12-11 02:41:00.385839+03	\N	CST	Central Standard Time (Chicago)	-06:00	America/Chicago	t
7	2025-12-11 02:41:00.429292+03	2025-12-11 02:41:00.429292+03	\N	MST	Mountain Standard Time (Denver)	-07:00	America/Denver	t
8	2025-12-11 02:41:00.486524+03	2025-12-11 02:41:00.486524+03	\N	PST	Pacific Standard Time (Los Angeles)	-08:00	America/Los_Angeles	t
9	2025-12-11 02:41:00.540054+03	2025-12-11 02:41:00.540054+03	\N	GMT	Greenwich Mean Time (London)	+00:00	Europe/London	t
10	2025-12-11 02:41:00.589638+03	2025-12-11 02:41:00.589638+03	\N	CET	Central European Time (Paris, Berlin)	+01:00	Europe/Paris	t
11	2025-12-11 02:41:00.628942+03	2025-12-11 02:41:00.628942+03	\N	EET	Eastern European Time (Athens, Cairo)	+02:00	Europe/Athens	t
12	2025-12-11 02:41:00.695545+03	2025-12-11 02:41:00.695545+03	\N	IST	India Standard Time (Mumbai, Delhi)	+05:30	Asia/Kolkata	t
13	2025-12-11 02:41:00.76586+03	2025-12-11 02:41:00.76586+03	\N	CST_CHINA	China Standard Time (Beijing, Shanghai)	+08:00	Asia/Shanghai	t
14	2025-12-11 02:41:00.829811+03	2025-12-11 02:41:00.829811+03	\N	JST	Japan Standard Time (Tokyo)	+09:00	Asia/Tokyo	t
15	2025-12-11 02:41:00.882363+03	2025-12-11 02:41:00.882363+03	\N	AEST	Australian Eastern Standard Time (Sydney)	+10:00	Australia/Sydney	t
16	2025-12-11 02:41:00.992927+03	2025-12-11 02:41:00.992927+03	\N	NZST	New Zealand Standard Time (Auckland)	+12:00	Pacific/Auckland	t
\.


--
-- Data for Name: two_factor_attempts; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.two_factor_attempts (id, created_at, updated_at, deleted_at, user_id, success, ip_address, user_agent, failure_type, attempted_at) FROM stdin;
1	2025-12-09 02:11:10.800946+03	2025-12-09 02:11:10.800946+03	\N	2	t	[::1]	PostmanRuntime/7.49.1	setup_completed	2025-12-09 02:11:10.768582+03
2	2025-12-09 02:18:17.912105+03	2025-12-09 02:18:17.912105+03	\N	2	t	[::1]	PostmanRuntime/7.49.1	login_verified	2025-12-09 02:18:17.886715+03
3	2025-12-09 02:42:50.420003+03	2025-12-09 02:42:50.420003+03	\N	2	f	[::1]	PostmanRuntime/7.49.1	invalid_code	2025-12-09 02:42:50.397173+03
4	2025-12-09 02:47:39.140393+03	2025-12-09 02:47:39.140393+03	\N	2	f	[::1]	PostmanRuntime/7.49.1	invalid_code	2025-12-09 02:47:39.095323+03
5	2025-12-09 02:54:47.018971+03	2025-12-09 02:54:47.018971+03	\N	2	t	[::1]	PostmanRuntime/7.49.1	setup_completed	2025-12-09 02:54:46.993191+03
6	2025-12-09 02:55:41.9336+03	2025-12-09 02:55:41.9336+03	\N	2	t	[::1]	PostmanRuntime/7.49.1	login_verified	2025-12-09 02:55:41.908472+03
\.


--
-- Data for Name: two_factor_auths; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.two_factor_auths (id, created_at, updated_at, deleted_at, user_id, enabled, secret, backup_codes_hash, verified_at, last_used_at, method) FROM stdin;
\.


--
-- Data for Name: two_factor_sessions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.two_factor_sessions (id, created_at, updated_at, deleted_at, user_id, secret, verified, expires_at, ip_address, user_agent) FROM stdin;
1	2025-12-09 01:51:10.105424+03	2025-12-09 01:51:10.105424+03	2025-12-09 02:09:21.960797+03	2	OMFS7F5CYZ2SS3UZXRCO6PDU2HHHB3YA	f	2025-12-09 02:06:10.047315+03	[::1]	PostmanRuntime/7.49.1
2	2025-12-09 02:09:22.011049+03	2025-12-09 02:09:22.011049+03	2025-12-09 02:11:10.678235+03	2	KGLUB4IKYKY3NI77N5PP2LOSV3KI3GRQ	f	2025-12-09 02:24:21.946946+03	[::1]	PostmanRuntime/7.49.1
3	2025-12-09 02:38:20.146306+03	2025-12-09 02:38:20.146306+03	2025-12-09 02:44:07.893189+03	2	RLGHGMO33Y6VBDACULH5TNLD4BV7UODS	f	2025-12-09 02:53:20.052544+03	[::1]	PostmanRuntime/7.49.1
4	2025-12-09 02:44:07.92546+03	2025-12-09 02:44:07.92546+03	2025-12-09 02:54:46.801982+03	2	5ALJKMSWLZSLSHYZIEILGFPMXO2CCSBX	f	2025-12-09 02:59:07.868666+03	[::1]	PostmanRuntime/7.49.1
\.


--
-- Data for Name: user_engagement_metrics; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.user_engagement_metrics (id, created_at, updated_at, deleted_at, account_id, date, session_start, session_end, session_duration, page_views, events_viewed, search_queries, tickets_purchased, events_bookmarked, social_shares, email_signups, revenue_generated, user_agent, ip_address, country, city, referrer_source, campaign_id, utm_source, utm_campaign, utm_medium) FROM stdin;
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (id, created_at, updated_at, deleted_at, account_id, first_name, last_name, username, phone, email, password, confirmation_code, isconfirmed, role, is_active, profile_picture, email_verified, email_verified_at, verification_token_exp, refresh_token, refresh_token_exp, last_login_at, token_version) FROM stdin;
2	2025-12-08 04:39:00.005752+03	2026-01-12 19:38:39.633644+03	\N	2	Kamau	Simon	kamausimon	0746960677	kamausimon217@gmail.com	$2a$12$DJLWduu/.Ixic99HHpak4uu/35Z2PQFTMGGmzRi0aPj7WsQgvrChu		f	customer	t	\N	t	2025-12-08 06:25:54.139277+03	\N	ea882bda310d2784ac6a313c09a84c736fdf660f5973735759ca509c498e48a2	1773419919	1768235919	1
9	2026-01-12 20:07:02.666148+03	2026-01-12 20:07:57.294769+03	\N	10	Dorcas	Gichuru	dorcaswairuri98	0746960671	dorcaswairuri98@gmail.com	$2a$12$kR4czVLvGjfjCAh9PAB7..hM53QE/tq7mc12.qb4MzU3JrRl8Da/C		f	customer	t	\N	f	\N	\N	9cede8ca2d5b8ae9b33d3e0fb3ef1aa6808f4c337cc8f9f8c9145dadd09215ad	1773421677	1768237677	1
6	2025-12-29 13:27:28.896193+03	2025-12-29 13:27:39.704821+03	\N	6	Test	User			testuser@example.com	$2a$12$2CqMyYKGlgUNQRR7tuAXhuIKhIJLYIeTdEUTB/9lzECIbHeA1LfRy		f	customer	t	\N	f	\N	\N	946800e89d3802763dc2d279d2205f09c9d9394761432cbd80b3c493f4bb21e5	1772188059	1767004059	1
10	2026-01-12 20:09:02.697076+03	2026-01-12 20:10:10.737896+03	\N	11	Dorcas	Gichuru	gichuruwairuri98	0746960672	gichuruwairuri98@gmail.com	$2a$12$c2EMt3LBLS1YrjTLicUOD.JL32ldAHatUU9umlPyCCw8bSs2/222e		f	customer	t	\N	f	\N	\N	277645da4b397851f0f5ccb35307c451e255990046c1dbe94efe2b89307175b7	1773421810	1768237810	1
4	2025-12-09 20:57:42.264524+03	2026-01-09 12:40:31.473423+03	\N	4	Admin	kamau	adminkamau	0729483698	topstonehelp@gmail.com	$2a$12$u/iBQs5C6CooCxocJElw2.o4afq.M99x/Up7hD43JBTzP/jgLSWF6		f	admin	t	\N	t	2025-12-09 20:58:19.965302+03	\N	a1dcecdd81b1408c4da76fceabe9c45db7b6bbc97d58477d32392150f60c703c	1773135631	1767951631	1
5	2025-12-10 01:27:18.05025+03	2025-12-10 01:39:44.354475+03	\N	5	Admin	Harsha	adminharsha	0729483697	harshasg10@gmail.com	$2a$12$rlREroakwOEu2jKoGfgKj.B1fZByG4WeA8RAQuLIDelSNtnjMJtk.		f	admin	t	\N	t	2025-12-10 01:27:50.18922+03	\N	ae58ae8a58dfd2178c5eb74bc9cd162d16e2593131e1b15d5c22ba552e4eaa0a	1770503293	1765319293	4
1	2025-12-03 00:27:40.866685+03	2025-12-11 00:18:20.909846+03	\N	1	Test	user	testuser	0712345678	test@example.com	$2a$12$q1q8D9.xpG3CeJhQahm1P.xoiBKwQFUTTMBeHxzlWVRIayePExB/e		f	customer	t	\N	f	\N	\N	0487c95e9567faa78e3d83962dd7bdd1c6252baa213c6cb991305629e923b166	1770585500	1765401500	1
3	2025-12-09 05:55:26.67715+03	2026-01-12 16:14:23.334507+03	\N	3	simon	kamau	simon	0799500565	topstonewriters@gmail.com	$2a$10$rs3eb90XnA5nnCMtTmiy5e1Rc7oj2co8SqkUjO9wgpFkqWY5nY6Ua		f	organizer	t	\N	t	2025-12-09 05:57:44.581342+03	\N	c8199fd4f70d2f637abfc8f331335775279391ec635242240716b98eb1a3a86b	1773407663	1768223663	1
\.


--
-- Data for Name: venues; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.venues (id, created_at, updated_at, deleted_at, venue_name, venue_capacity, venue_section, venue_type, venue_location, address, city, state, country, zip_code, parking_available, parking_capacity, is_accessible, has_wifi, has_catering, contact_email, contact_phone, website) FROM stdin;
\.


--
-- Data for Name: waitlist_entries; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.waitlist_entries (id, created_at, updated_at, deleted_at, event_id, ticket_class_id, email, name, phone, quantity, status, notified_at, converted_at, expires_at, priority, session_id, user_id) FROM stdin;
1	2025-12-28 09:17:17.037095+03	2025-12-28 10:10:39.36823+03	\N	2	\N	Kamausimon217@gmail.com	simon kamau	0746960677	2	expired	\N	\N	\N	0		\N
\.


--
-- Data for Name: webhook_logs; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.webhook_logs (id, created_at, updated_at, deleted_at, provider, event_id, event_type, status, payload, headers, request_method, request_path, processed_at, processing_time, retry_count, last_retry_at, success, error_message, stack_trace, order_id, payment_transaction_id, payment_record_id, account_id, organizer_id, external_transaction_id, external_reference, signature_valid, signature_header, ip_address, user_agent, idempotency_key, is_duplicate, environment, api_version, notes, response_status, response_body) FROM stdin;
1	2025-12-15 06:00:08.662251+03	2025-12-15 06:00:08.662251+03	\N	intasend	unknown	payment	failed	{"test":"connection"}	{"Accept":["*/*"],"Accept-Encoding":["gzip"],"Content-Length":["21"],"Content-Type":["application/json"],"User-Agent":["curl/8.5.0"],"X-Forwarded-For":["102.209.76.114"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https"]}	POST		2025-12-15 06:00:08.609302+03	\N	0	\N	f	Invalid signature	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	127.0.0.1:42896	curl/8.5.0	\N	f	production	\N	\N	200	\N
2	2025-12-15 06:01:12.106261+03	2025-12-15 06:01:12.106261+03	\N	intasend	unknown	payment	failed	{"test":"connection"}	{"Accept":["*/*"],"Accept-Encoding":["gzip"],"Content-Length":["21"],"Content-Type":["application/json"],"User-Agent":["curl/8.5.0"],"X-Forwarded-For":["102.209.76.114"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https"]}	POST		2025-12-15 06:01:12.074498+03	\N	0	\N	f	Invalid signature	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	127.0.0.1:40720	curl/8.5.0	\N	f	production	\N	\N	200	\N
3	2025-12-15 06:02:34.306211+03	2025-12-15 06:02:34.306211+03	\N	intasend	unknown	payment	failed	{"test":"connection"}	{"Accept":["*/*"],"Accept-Encoding":["gzip"],"Content-Length":["21"],"Content-Type":["application/json"],"User-Agent":["curl/8.5.0"],"X-Forwarded-For":["102.209.76.114"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https"]}	POST		2025-12-15 06:02:20.352298+03	\N	0	\N	f	Invalid signature	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	127.0.0.1:55244	curl/8.5.0	\N	f	production	\N	\N	200	\N
4	2025-12-15 06:06:08.088104+03	2025-12-15 06:06:08.088104+03	\N	intasend	unknown	payment	failed	{"test":"connection"}	{"Accept":["*/*"],"Accept-Encoding":["gzip"],"Content-Length":["21"],"Content-Type":["application/json"],"User-Agent":["curl/8.5.0"],"X-Forwarded-For":["102.209.76.114"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https"]}	POST		2025-12-15 06:06:08.0663+03	\N	0	\N	f	Invalid signature	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	127.0.0.1:53038	curl/8.5.0	\N	f	production	\N	\N	200	\N
5	2025-12-15 06:12:16.603815+03	2025-12-15 06:12:16.603815+03	\N	intasend	unknown	payment	failed	{"test":"connection"}	{"Accept":["*/*"],"Accept-Encoding":["gzip"],"Content-Length":["21"],"Content-Type":["application/json"],"User-Agent":["curl/8.5.0"],"X-Forwarded-For":["102.209.76.114"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https"]}	POST		2025-12-15 06:12:16.590199+03	\N	0	\N	f	Invalid signature	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	127.0.0.1:40926	curl/8.5.0	\N	f	production	\N	\N	200	\N
6	2025-12-15 06:13:04.833628+03	2025-12-15 06:13:04.833628+03	\N	intasend	unknown	payment	failed	{"invoice_id": "QJOD8DR", "state": "PENDING", "provider": "M-PESA", "charges": "0.00", "net_amount": "10.00", "currency": "KES", "value": "10.00", "account": "254746960677", "api_ref": "Direct Bill", "mpesa_reference": null, "host": "https://sandbox.intasend.com", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-15T06:13:03.765979+03:00", "updated_at": "2025-12-15T06:13:03.790972+03:00", "challenge": "Kamau0746"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=cd0b633686f548749b7f443ad728c2a9,sentry-environment=sandbox,sentry-release=3f5eef9c1523ae2b4cdfc2f912408cd824fb1d6b,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["533"],"Content-Type":["application/json"],"Sentry-Trace":["cd0b633686f548749b7f443ad728c2a9-81ca82e30cdb823d"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-15 06:13:04.820531+03	\N	0	\N	f	Invalid signature	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	127.0.0.1:45976	_	\N	f	production	\N	\N	200	\N
7	2025-12-15 06:13:11.392608+03	2025-12-15 06:13:11.392608+03	\N	intasend	unknown	payment	failed	{"invoice_id": "QJOD8DR", "state": "PROCESSING", "provider": "M-PESA", "charges": "0.00", "net_amount": "10.00", "currency": "KES", "value": "10.00", "account": "254746960677", "api_ref": "Direct Bill", "mpesa_reference": null, "host": "https://sandbox.intasend.com", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-15T06:13:03.765979+03:00", "updated_at": "2025-12-15T06:13:10.513555+03:00", "challenge": "Kamau0746"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=cd0b633686f548749b7f443ad728c2a9,sentry-environment=sandbox,sentry-release=3f5eef9c1523ae2b4cdfc2f912408cd824fb1d6b,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["536"],"Content-Type":["application/json"],"Sentry-Trace":["cd0b633686f548749b7f443ad728c2a9-805c629c0451c486"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-15 06:13:11.379719+03	\N	0	\N	f	Invalid signature	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	127.0.0.1:45986	_	\N	f	production	\N	\N	200	\N
8	2025-12-15 06:15:10.734716+03	2025-12-15 06:15:10.734716+03	\N	intasend	unknown	payment	failed	{"test":"connection"}	{"Accept":["*/*"],"Accept-Encoding":["gzip"],"Content-Length":["21"],"Content-Type":["application/json"],"User-Agent":["curl/8.5.0"],"X-Forwarded-For":["102.209.76.114"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https"]}	POST		2025-12-15 06:15:10.721216+03	\N	0	\N	f	Invalid signature	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	127.0.0.1:54362	curl/8.5.0	\N	f	production	\N	\N	200	\N
9	2025-12-15 06:31:15.37593+03	2025-12-15 06:31:15.37593+03	\N	intasend	unknown	payment	failed	{"invoice_id": "RKP72EY", "state": "PENDING", "provider": "M-PESA", "charges": "0.00", "net_amount": "3.00", "currency": "KES", "value": "3.00", "account": "254708374149", "api_ref": "ORD-1-1765769473", "mpesa_reference": null, "host": "102.209.76.114", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-15T06:31:14.286323+03:00", "updated_at": "2025-12-15T06:31:14.301762+03:00", "challenge": "Kamau0746"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=90761ca181154df1a0660e4009c6a44d,sentry-environment=sandbox,sentry-release=3f5eef9c1523ae2b4cdfc2f912408cd824fb1d6b,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["522"],"Content-Type":["application/json"],"Sentry-Trace":["90761ca181154df1a0660e4009c6a44d-836c010b997885f4"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-15 06:31:15.362447+03	\N	0	\N	f	Invalid signature	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	127.0.0.1:42172	_	\N	f	production	\N	\N	200	\N
10	2025-12-15 06:31:21.703972+03	2025-12-15 06:31:21.703972+03	\N	intasend	unknown	payment	failed	{"invoice_id": "RKP72EY", "state": "PROCESSING", "provider": "M-PESA", "charges": "0.00", "net_amount": "3.00", "currency": "KES", "value": "3.00", "account": "254708374149", "api_ref": "ORD-1-1765769473", "mpesa_reference": null, "host": "102.209.76.114", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-15T06:31:14.286323+03:00", "updated_at": "2025-12-15T06:31:20.791756+03:00", "challenge": "Kamau0746"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=90761ca181154df1a0660e4009c6a44d,sentry-environment=sandbox,sentry-release=3f5eef9c1523ae2b4cdfc2f912408cd824fb1d6b,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["525"],"Content-Type":["application/json"],"Sentry-Trace":["90761ca181154df1a0660e4009c6a44d-be505bd33f482086"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-15 06:31:21.688283+03	\N	0	\N	f	Invalid signature	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	127.0.0.1:42174	_	\N	f	production	\N	\N	200	\N
11	2025-12-15 06:40:34.759738+03	2025-12-15 06:40:34.759738+03	\N	intasend	unknown	payment	failed	{"invoice_id": "R8PGNVR", "state": "PENDING", "provider": "M-PESA", "charges": "0.00", "net_amount": "3.00", "currency": "KES", "value": "3.00", "account": "254708374149", "api_ref": "ORD-1-1765770033", "mpesa_reference": null, "host": "102.209.76.114", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-15T06:40:33.745375+03:00", "updated_at": "2025-12-15T06:40:33.759995+03:00", "challenge": "Kamau0746"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=a2eb12b610334ecda1dbb6c658745cde,sentry-environment=sandbox,sentry-release=3f5eef9c1523ae2b4cdfc2f912408cd824fb1d6b,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["522"],"Content-Type":["application/json"],"Sentry-Trace":["a2eb12b610334ecda1dbb6c658745cde-ac81d964b1c58aa1"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-15 06:40:34.739637+03	\N	0	\N	f	Invalid signature	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	127.0.0.1:44998	_	\N	f	production	\N	\N	200	\N
12	2025-12-15 06:40:41.062534+03	2025-12-15 06:40:41.062534+03	\N	intasend	unknown	payment	failed	{"invoice_id": "R8PGNVR", "state": "PROCESSING", "provider": "M-PESA", "charges": "0.00", "net_amount": "3.00", "currency": "KES", "value": "3.00", "account": "254708374149", "api_ref": "ORD-1-1765770033", "mpesa_reference": null, "host": "102.209.76.114", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-15T06:40:33.745375+03:00", "updated_at": "2025-12-15T06:40:40.144360+03:00", "challenge": "Kamau0746"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=a2eb12b610334ecda1dbb6c658745cde,sentry-environment=sandbox,sentry-release=3f5eef9c1523ae2b4cdfc2f912408cd824fb1d6b,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["525"],"Content-Type":["application/json"],"Sentry-Trace":["a2eb12b610334ecda1dbb6c658745cde-842e72d28e85777c"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-15 06:40:41.042519+03	\N	0	\N	f	Invalid signature	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	127.0.0.1:45002	_	\N	f	production	\N	\N	200	\N
13	2025-12-15 06:44:47.646927+03	2025-12-15 06:44:47.646927+03	\N	intasend	unknown	payment	failed	{"invoice_id": "QP6E8ZY", "state": "PENDING", "provider": "M-PESA", "charges": "0.00", "net_amount": "3.00", "currency": "KES", "value": "3.00", "account": "254708374149", "api_ref": "ORD-1-1765770286", "mpesa_reference": null, "host": "102.209.76.114", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-15T06:44:46.604355+03:00", "updated_at": "2025-12-15T06:44:46.618579+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=31f3579741924999b16adf4392fdd279,sentry-environment=sandbox,sentry-release=3f5eef9c1523ae2b4cdfc2f912408cd824fb1d6b,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["528"],"Content-Type":["application/json"],"Sentry-Trace":["31f3579741924999b16adf4392fdd279-92068ac1e84c9ec1"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-15 06:44:47.623956+03	\N	0	\N	f	Invalid signature	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	127.0.0.1:57814	_	\N	f	production	\N	\N	200	\N
14	2025-12-15 06:44:53.930709+03	2025-12-15 06:44:53.930709+03	\N	intasend	unknown	payment	failed	{"invoice_id": "QP6E8ZY", "state": "PROCESSING", "provider": "M-PESA", "charges": "0.00", "net_amount": "3.00", "currency": "KES", "value": "3.00", "account": "254708374149", "api_ref": "ORD-1-1765770286", "mpesa_reference": null, "host": "102.209.76.114", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-15T06:44:46.604355+03:00", "updated_at": "2025-12-15T06:44:52.986303+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=31f3579741924999b16adf4392fdd279,sentry-environment=sandbox,sentry-release=3f5eef9c1523ae2b4cdfc2f912408cd824fb1d6b,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["531"],"Content-Type":["application/json"],"Sentry-Trace":["31f3579741924999b16adf4392fdd279-8e6890da00a5169b"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-15 06:44:53.913491+03	\N	0	\N	f	Invalid signature	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	127.0.0.1:34686	_	\N	f	production	\N	\N	200	\N
15	2025-12-15 06:48:46.083374+03	2025-12-15 06:48:46.083374+03	\N	intasend	unknown	payment	failed	{"invoice_id": "R0Z8VGR", "state": "PENDING", "provider": "M-PESA", "charges": "0.00", "net_amount": "3.00", "currency": "KES", "value": "3.00", "account": "254708374149", "api_ref": "ORD-1-1765770524", "mpesa_reference": null, "host": "102.209.76.114", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-15T06:48:45.009668+03:00", "updated_at": "2025-12-15T06:48:45.028711+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=97bb776a07634fbba389b45de1f664b3,sentry-environment=sandbox,sentry-release=3f5eef9c1523ae2b4cdfc2f912408cd824fb1d6b,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["528"],"Content-Type":["application/json"],"Sentry-Trace":["97bb776a07634fbba389b45de1f664b3-9a8b5a443651702a"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-15 06:48:46.06489+03	\N	0	\N	f	Invalid signature	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	127.0.0.1:39088	_	\N	f	production	\N	\N	200	\N
16	2025-12-15 06:48:52.389895+03	2025-12-15 06:48:52.389895+03	\N	intasend	unknown	payment	failed	{"invoice_id": "R0Z8VGR", "state": "PROCESSING", "provider": "M-PESA", "charges": "0.00", "net_amount": "3.00", "currency": "KES", "value": "3.00", "account": "254708374149", "api_ref": "ORD-1-1765770524", "mpesa_reference": null, "host": "102.209.76.114", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-15T06:48:45.009668+03:00", "updated_at": "2025-12-15T06:48:51.470552+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=97bb776a07634fbba389b45de1f664b3,sentry-environment=sandbox,sentry-release=3f5eef9c1523ae2b4cdfc2f912408cd824fb1d6b,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["531"],"Content-Type":["application/json"],"Sentry-Trace":["97bb776a07634fbba389b45de1f664b3-87762e8e05afab37"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-15 06:48:52.375785+03	\N	0	\N	f	Invalid signature	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	127.0.0.1:50380	_	\N	f	production	\N	\N	200	\N
17	2025-12-15 06:51:23.574215+03	2025-12-15 06:51:23.574215+03	\N	intasend	unknown	payment	failed	{"invoice_id": "RG6ZWMQ", "state": "PENDING", "provider": "M-PESA", "charges": "0.00", "net_amount": "3.00", "currency": "KES", "value": "3.00", "account": "254708374149", "api_ref": "ORD-1-1765770681", "mpesa_reference": null, "host": "102.209.76.114", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-15T06:51:22.387890+03:00", "updated_at": "2025-12-15T06:51:22.401077+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=ff14006dd7f24283af37b3510143ae6e,sentry-environment=sandbox,sentry-release=3f5eef9c1523ae2b4cdfc2f912408cd824fb1d6b,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["528"],"Content-Type":["application/json"],"Sentry-Trace":["ff14006dd7f24283af37b3510143ae6e-bfdf0d1013714c80"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-15 06:51:23.548813+03	\N	0	\N	f	Failed to parse JSON: json: cannot unmarshal string into Go struct field IntasendWebhookEvent.charges of type float64	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	127.0.0.1:38270	_	\N	f	production	\N	\N	200	\N
18	2025-12-15 06:51:29.575378+03	2025-12-15 06:51:29.575378+03	\N	intasend	unknown	payment	failed	{"invoice_id": "RG6ZWMQ", "state": "PROCESSING", "provider": "M-PESA", "charges": "0.00", "net_amount": "3.00", "currency": "KES", "value": "3.00", "account": "254708374149", "api_ref": "ORD-1-1765770681", "mpesa_reference": null, "host": "102.209.76.114", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-15T06:51:22.387890+03:00", "updated_at": "2025-12-15T06:51:28.677634+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=ff14006dd7f24283af37b3510143ae6e,sentry-environment=sandbox,sentry-release=3f5eef9c1523ae2b4cdfc2f912408cd824fb1d6b,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["531"],"Content-Type":["application/json"],"Sentry-Trace":["ff14006dd7f24283af37b3510143ae6e-999c2cf1a384fa50"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-15 06:51:29.544275+03	\N	0	\N	f	Failed to parse JSON: json: cannot unmarshal string into Go struct field IntasendWebhookEvent.charges of type float64	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	127.0.0.1:38278	_	\N	f	production	\N	\N	200	\N
19	2025-12-15 06:56:06.634036+03	2025-12-15 06:56:06.634036+03	\N	intasend		payment	processed	{"invoice_id": "RONXG2Y", "state": "PENDING", "provider": "M-PESA", "charges": "0.00", "net_amount": "3.00", "currency": "KES", "value": "3.00", "account": "254708374149", "api_ref": "ORD-1-1765770964", "mpesa_reference": null, "host": "102.209.76.114", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-15T06:56:05.203786+03:00", "updated_at": "2025-12-15T06:56:05.224045+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=17fd1af311a1462eaf1e90386dda0ee2,sentry-environment=sandbox,sentry-release=3f5eef9c1523ae2b4cdfc2f912408cd824fb1d6b,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["528"],"Content-Type":["application/json"],"Sentry-Trace":["17fd1af311a1462eaf1e90386dda0ee2-8e2c93a5b91f7778"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-15 06:56:06.61712+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	\N	t	\N	127.0.0.1:57986	_	\N	f	production	\N	\N	200	\N
20	2025-12-15 06:56:12.884109+03	2025-12-15 06:56:12.884109+03	\N	intasend		payment	processed	{"invoice_id": "RONXG2Y", "state": "PROCESSING", "provider": "M-PESA", "charges": "0.00", "net_amount": "3.00", "currency": "KES", "value": "3.00", "account": "254708374149", "api_ref": "ORD-1-1765770964", "mpesa_reference": null, "host": "102.209.76.114", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-15T06:56:05.203786+03:00", "updated_at": "2025-12-15T06:56:11.868508+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=17fd1af311a1462eaf1e90386dda0ee2,sentry-environment=sandbox,sentry-release=3f5eef9c1523ae2b4cdfc2f912408cd824fb1d6b,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["531"],"Content-Type":["application/json"],"Sentry-Trace":["17fd1af311a1462eaf1e90386dda0ee2-ad390ed650ab4d80"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-15 06:56:12.862468+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	\N	t	\N	127.0.0.1:50706	_	\N	f	production	\N	\N	200	\N
21	2025-12-15 07:03:44.433898+03	2025-12-15 07:03:44.433898+03	\N	intasend		payment	processed	{"invoice_id": "YV56VGY", "state": "PENDING", "provider": "M-PESA", "charges": "0.00", "net_amount": "10.00", "currency": "KES", "value": "10.00", "account": "254746960677", "api_ref": "ORD-1-1765771423", "mpesa_reference": null, "host": "102.209.76.114", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-15T07:03:43.348242+03:00", "updated_at": "2025-12-15T07:03:43.364708+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=2a5c40c1b4014f74b2e689c0fce1fdd7,sentry-environment=sandbox,sentry-release=3f5eef9c1523ae2b4cdfc2f912408cd824fb1d6b,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["530"],"Content-Type":["application/json"],"Sentry-Trace":["2a5c40c1b4014f74b2e689c0fce1fdd7-941eda5cf7be94b6"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-15 07:03:44.41189+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	\N	t	\N	127.0.0.1:42096	_	\N	f	production	\N	\N	200	\N
22	2025-12-15 07:03:50.368838+03	2025-12-15 07:03:50.368838+03	\N	intasend		payment	processed	{"invoice_id": "YV56VGY", "state": "PROCESSING", "provider": "M-PESA", "charges": "0.00", "net_amount": "10.00", "currency": "KES", "value": "10.00", "account": "254746960677", "api_ref": "ORD-1-1765771423", "mpesa_reference": null, "host": "102.209.76.114", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-15T07:03:43.348242+03:00", "updated_at": "2025-12-15T07:03:49.459730+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=2a5c40c1b4014f74b2e689c0fce1fdd7,sentry-environment=sandbox,sentry-release=3f5eef9c1523ae2b4cdfc2f912408cd824fb1d6b,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["533"],"Content-Type":["application/json"],"Sentry-Trace":["2a5c40c1b4014f74b2e689c0fce1fdd7-95f1b7815a5b53ae"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-15 07:03:50.34314+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	\N	t	\N	127.0.0.1:32884	_	\N	f	production	\N	\N	200	\N
23	2025-12-15 07:04:35.653577+03	2025-12-15 07:04:35.653577+03	\N	intasend		payment	processed	{"invoice_id": "Q3BO3EQ", "state": "PENDING", "provider": "M-PESA", "charges": "0.00", "net_amount": "10.00", "currency": "KES", "value": "10.00", "account": "254799500565", "api_ref": "ORD-1-1765771474", "mpesa_reference": null, "host": "102.209.76.114", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-15T07:04:34.568691+03:00", "updated_at": "2025-12-15T07:04:34.582552+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=5272e256c10648fd87207ca0fdf02248,sentry-environment=sandbox,sentry-release=3f5eef9c1523ae2b4cdfc2f912408cd824fb1d6b,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["530"],"Content-Type":["application/json"],"Sentry-Trace":["5272e256c10648fd87207ca0fdf02248-942fd6635c23478c"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-15 07:04:35.630361+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	\N	t	\N	127.0.0.1:53184	_	\N	f	production	\N	\N	200	\N
24	2025-12-15 07:04:55.705577+03	2025-12-15 07:04:55.705577+03	\N	intasend		payment	processed	{"invoice_id": "Q3BO3EQ", "state": "PROCESSING", "provider": "M-PESA", "charges": "0.00", "net_amount": "10.00", "currency": "KES", "value": "10.00", "account": "254799500565", "api_ref": "ORD-1-1765771474", "mpesa_reference": null, "host": "102.209.76.114", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-15T07:04:34.568691+03:00", "updated_at": "2025-12-15T07:04:40.893871+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=5272e256c10648fd87207ca0fdf02248,sentry-environment=sandbox,sentry-release=3f5eef9c1523ae2b4cdfc2f912408cd824fb1d6b,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["533"],"Content-Type":["application/json"],"Sentry-Trace":["5272e256c10648fd87207ca0fdf02248-be2a876a52d236bd"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-15 07:04:55.686121+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	\N	t	\N	127.0.0.1:37634	_	\N	f	production	\N	\N	200	\N
25	2025-12-16 05:44:21.258912+03	2025-12-16 05:44:21.258912+03	\N	intasend		payment	processed	{"invoice_id": "REK4ZGR", "state": "PENDING", "provider": "CARD-PAYMENT", "charges": "0.00", "net_amount": "2.48", "currency": "KES", "value": "2.48", "account": "kamausimon217@gmail.com", "api_ref": "ORD-1-1765852997", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T05:44:19.498245+03:00", "updated_at": "2025-12-16T05:44:19.516114+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=12bb15806ecc4829bd71472e791b2b50,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["540"],"Content-Type":["application/json"],"Sentry-Trace":["12bb15806ecc4829bd71472e791b2b50-aac4dc720f94cb60"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 05:44:21.17132+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	\N	t	\N	127.0.0.1:39032	_	\N	f	production	\N	\N	200	\N
26	2025-12-16 05:44:37.426241+03	2025-12-16 05:44:37.426241+03	\N	intasend		payment	processed	{"invoice_id": "REK4ZGR", "state": "PROCESSING", "provider": "CARD-PAYMENT", "charges": "0.00", "net_amount": "2.48", "currency": "KES", "value": "2.48", "account": "kamausimon217@gmail.com", "api_ref": "ORD-1-1765852997", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T05:44:19.498245+03:00", "updated_at": "2025-12-16T05:44:21.033049+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=12bb15806ecc4829bd71472e791b2b50,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["543"],"Content-Type":["application/json"],"Sentry-Trace":["12bb15806ecc4829bd71472e791b2b50-8eb29787e5eacbff"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 05:44:37.395388+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	\N	t	\N	127.0.0.1:36994	_	\N	f	production	\N	\N	200	\N
27	2025-12-16 05:44:33.84729+03	2025-12-16 05:44:33.84729+03	\N	intasend		payment	processed	{"invoice_id": "REK4ZGR", "state": "COMPLETE", "provider": "CARD-PAYMENT", "charges": "0.09", "net_amount": "2.39", "currency": "KES", "value": "2.48", "account": "kamausimon217@gmail.com", "api_ref": "ORD-1-1765852997", "clearing_status": "CLEARING", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": "MASTERCARD"}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T05:44:19.498245+03:00", "updated_at": "2025-12-16T05:44:32.692661+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=c8bcfbf1dbe84b578667b8a9fcc7ad43,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["580"],"Content-Type":["application/json"],"Sentry-Trace":["c8bcfbf1dbe84b578667b8a9fcc7ad43-b95bbb9da5520020"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 05:44:33.639653+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	\N	t	\N	127.0.0.1:40952	_	\N	f	production	\N	\N	200	\N
28	2025-12-16 06:00:36.80066+03	2025-12-16 06:00:36.80066+03	\N	intasend		payment	processed	{"invoice_id": "Y4N8VEQ", "state": "PENDING", "provider": "CARD-PAYMENT", "charges": "0.00", "net_amount": "2.48", "currency": "KES", "value": "2.48", "account": "kamausimon217@gmail.com", "api_ref": "ORD-1-1765853986", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T06:00:35.529741+03:00", "updated_at": "2025-12-16T06:00:35.551989+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=f59bf9ae98a148199ef5c84854c3937b,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["540"],"Content-Type":["application/json"],"Sentry-Trace":["f59bf9ae98a148199ef5c84854c3937b-9084153faa9af529"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 06:00:36.77751+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	\N	t	\N	127.0.0.1:36122	_	\N	f	production	\N	\N	200	\N
29	2025-12-16 06:00:37.888743+03	2025-12-16 06:00:37.888743+03	\N	intasend		payment	processed	{"invoice_id": "Y4N8VEQ", "state": "PROCESSING", "provider": "CARD-PAYMENT", "charges": "0.00", "net_amount": "2.48", "currency": "KES", "value": "2.48", "account": "kamausimon217@gmail.com", "api_ref": "ORD-1-1765853986", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T06:00:35.529741+03:00", "updated_at": "2025-12-16T06:00:36.932931+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=f59bf9ae98a148199ef5c84854c3937b,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["543"],"Content-Type":["application/json"],"Sentry-Trace":["f59bf9ae98a148199ef5c84854c3937b-a336cff89d257ec3"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 06:00:37.864623+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	\N	t	\N	127.0.0.1:36134	_	\N	f	production	\N	\N	200	\N
30	2025-12-16 06:01:34.42974+03	2025-12-16 06:01:34.42974+03	\N	intasend		payment	processed	{"invoice_id": "Y4N8VEQ", "state": "FAILED", "provider": "CARD-PAYMENT", "charges": "0.00", "net_amount": "2.48", "currency": "KES", "value": "2.48", "account": "kamausimon217@gmail.com", "api_ref": "ORD-1-1765853986", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": "VISA"}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T06:00:35.529741+03:00", "updated_at": "2025-12-16T06:01:18.147617+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=4ecb576ebbab406cbcbb84357509987d,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["541"],"Content-Type":["application/json"],"Sentry-Trace":["4ecb576ebbab406cbcbb84357509987d-be342c9f74c66c89"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 06:01:34.380779+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	\N	t	\N	127.0.0.1:35086	_	\N	f	production	\N	\N	200	\N
31	2025-12-16 06:01:38.173638+03	2025-12-16 06:01:38.173638+03	\N	intasend		payment	processed	{"invoice_id": "RLPB58Q", "state": "PENDING", "provider": "CARD-PAYMENT", "charges": "0.00", "net_amount": "2.48", "currency": "KES", "value": "2.48", "account": "kamausimon217@gmail.com", "api_ref": "ORD-1-1765853986", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T06:01:21.907802+03:00", "updated_at": "2025-12-16T06:01:21.926903+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=f5fe868273024c048aebeb0b10863853,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["540"],"Content-Type":["application/json"],"Sentry-Trace":["f5fe868273024c048aebeb0b10863853-9ad8358557cf6453"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 06:01:38.139573+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	\N	t	\N	127.0.0.1:57172	_	\N	f	production	\N	\N	200	\N
32	2025-12-16 06:01:39.421126+03	2025-12-16 06:01:39.421126+03	\N	intasend		payment	processed	{"invoice_id": "RLPB58Q", "state": "PROCESSING", "provider": "CARD-PAYMENT", "charges": "0.00", "net_amount": "2.48", "currency": "KES", "value": "2.48", "account": "kamausimon217@gmail.com", "api_ref": "ORD-1-1765853986", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T06:01:21.907802+03:00", "updated_at": "2025-12-16T06:01:23.192571+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=f5fe868273024c048aebeb0b10863853,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["543"],"Content-Type":["application/json"],"Sentry-Trace":["f5fe868273024c048aebeb0b10863853-b563819a1914e710"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 06:01:39.39773+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	\N	t	\N	127.0.0.1:57188	_	\N	f	production	\N	\N	200	\N
33	2025-12-16 06:01:32.603749+03	2025-12-16 06:01:32.603749+03	\N	intasend		payment	processed	{"invoice_id": "RLPB58Q", "state": "COMPLETE", "provider": "CARD-PAYMENT", "charges": "0.09", "net_amount": "2.39", "currency": "KES", "value": "2.48", "account": "kamausimon217@gmail.com", "api_ref": "ORD-1-1765853986", "clearing_status": "CLEARING", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": "MASTERCARD"}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T06:01:21.907802+03:00", "updated_at": "2025-12-16T06:01:31.610576+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=a49c5b4fff474c5f8d2f337025ed2cc3,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["580"],"Content-Type":["application/json"],"Sentry-Trace":["a49c5b4fff474c5f8d2f337025ed2cc3-b63c338256c803f7"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 06:01:32.54158+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	\N	t	\N	127.0.0.1:49560	_	\N	f	production	\N	\N	200	\N
34	2025-12-16 06:16:59.664188+03	2025-12-16 06:16:59.664188+03	\N	intasend		payment	processed	{"invoice_id": "RZ364ZR", "state": "PENDING", "provider": "CARD-PAYMENT", "charges": "0.00", "net_amount": "2.48", "currency": "KES", "value": "2.48", "account": "kamausimon217@gmail.com", "api_ref": "ORD-1-1765855003", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T06:16:58.385317+03:00", "updated_at": "2025-12-16T06:16:58.411709+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=997fe6fbb9eb489a94d6dc7846f62f01,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["540"],"Content-Type":["application/json"],"Sentry-Trace":["997fe6fbb9eb489a94d6dc7846f62f01-9ec15dbb587e0aa5"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 06:16:59.619708+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	RZ364ZR	t	\N	127.0.0.1:53300	_	\N	f	production	\N	\N	200	\N
35	2025-12-16 06:17:00.742528+03	2025-12-16 06:17:00.742528+03	\N	intasend		payment	processed	{"invoice_id": "RZ364ZR", "state": "PROCESSING", "provider": "CARD-PAYMENT", "charges": "0.00", "net_amount": "2.48", "currency": "KES", "value": "2.48", "account": "kamausimon217@gmail.com", "api_ref": "ORD-1-1765855003", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T06:16:58.385317+03:00", "updated_at": "2025-12-16T06:16:59.668650+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=997fe6fbb9eb489a94d6dc7846f62f01,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["543"],"Content-Type":["application/json"],"Sentry-Trace":["997fe6fbb9eb489a94d6dc7846f62f01-ac3a2aafae49f3b9"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 06:17:00.716476+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	RZ364ZR	t	\N	127.0.0.1:53310	_	\N	f	production	\N	\N	200	\N
36	2025-12-16 06:17:11.006385+03	2025-12-16 06:17:11.006385+03	\N	intasend		payment	processed	{"invoice_id": "RZ364ZR", "state": "COMPLETE", "provider": "CARD-PAYMENT", "charges": "0.09", "net_amount": "2.39", "currency": "KES", "value": "2.48", "account": "kamausimon217@gmail.com", "api_ref": "ORD-1-1765855003", "clearing_status": "CLEARING", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": "MASTERCARD"}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T06:16:58.385317+03:00", "updated_at": "2025-12-16T06:17:09.716764+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=6428964083714d38aa882b80d8b65bb4,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["580"],"Content-Type":["application/json"],"Sentry-Trace":["6428964083714d38aa882b80d8b65bb4-af73ed71c4a8ee2c"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 06:17:10.986337+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	RZ364ZR	t	\N	127.0.0.1:38188	_	\N	f	production	\N	\N	200	\N
37	2025-12-16 06:21:57.627766+03	2025-12-16 06:21:57.627766+03	\N	intasend		payment	processed	{"invoice_id": "R5W822Q", "state": "PENDING", "provider": "CARD-PAYMENT", "charges": "0.00", "net_amount": "2.48", "currency": "KES", "value": "2.48", "account": "kamausimon217@gmail.com", "api_ref": "ORD-1-1765855297", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T06:21:56.347362+03:00", "updated_at": "2025-12-16T06:21:56.361250+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=ef1e0aa4dafb4352b108a2ee7f2fcc18,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["540"],"Content-Type":["application/json"],"Sentry-Trace":["ef1e0aa4dafb4352b108a2ee7f2fcc18-a59cfd8cbde692fd"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 06:21:57.593405+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	R5W822Q	t	\N	127.0.0.1:39698	_	\N	f	production	\N	\N	200	\N
38	2025-12-16 06:21:58.610806+03	2025-12-16 06:21:58.610806+03	\N	intasend		payment	processed	{"invoice_id": "R5W822Q", "state": "PROCESSING", "provider": "CARD-PAYMENT", "charges": "0.00", "net_amount": "2.48", "currency": "KES", "value": "2.48", "account": "kamausimon217@gmail.com", "api_ref": "ORD-1-1765855297", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T06:21:56.347362+03:00", "updated_at": "2025-12-16T06:21:57.561133+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=ef1e0aa4dafb4352b108a2ee7f2fcc18,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["543"],"Content-Type":["application/json"],"Sentry-Trace":["ef1e0aa4dafb4352b108a2ee7f2fcc18-a291540ad53db264"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 06:21:58.574287+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	R5W822Q	t	\N	127.0.0.1:39714	_	\N	f	production	\N	\N	200	\N
39	2025-12-16 06:22:07.004113+03	2025-12-16 06:22:07.004113+03	\N	intasend		payment	processed	{"invoice_id": "R5W822Q", "state": "COMPLETE", "provider": "CARD-PAYMENT", "charges": "0.09", "net_amount": "2.39", "currency": "KES", "value": "2.48", "account": "kamausimon217@gmail.com", "api_ref": "ORD-1-1765855297", "clearing_status": "CLEARING", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": "MASTERCARD"}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T06:21:56.347362+03:00", "updated_at": "2025-12-16T06:22:05.667149+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=f5ff7e146e454e458a672ba5319d12fd,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["580"],"Content-Type":["application/json"],"Sentry-Trace":["f5ff7e146e454e458a672ba5319d12fd-ab36de53174baefa"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 06:22:06.978621+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	R5W822Q	t	\N	127.0.0.1:47192	_	\N	f	production	\N	\N	200	\N
40	2025-12-16 06:25:02.412492+03	2025-12-16 06:25:02.412492+03	\N	intasend		payment	processed	{"invoice_id": "Q2N8JLY", "state": "PENDING", "provider": "CARD-PAYMENT", "charges": "0.00", "net_amount": "2.48", "currency": "KES", "value": "2.48", "account": "kamausimon217@gmail.com", "api_ref": "ORD-1-1765855490", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T06:25:01.230468+03:00", "updated_at": "2025-12-16T06:25:01.253563+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=1843b005877844c3867a9020ec15e636,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["540"],"Content-Type":["application/json"],"Sentry-Trace":["1843b005877844c3867a9020ec15e636-bda6143e48746d44"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 06:25:02.382507+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	Q2N8JLY	t	\N	127.0.0.1:41098	_	\N	f	production	\N	\N	200	\N
41	2025-12-16 06:25:03.645182+03	2025-12-16 06:25:03.645182+03	\N	intasend		payment	processed	{"invoice_id": "Q2N8JLY", "state": "PROCESSING", "provider": "CARD-PAYMENT", "charges": "0.00", "net_amount": "2.48", "currency": "KES", "value": "2.48", "account": "kamausimon217@gmail.com", "api_ref": "ORD-1-1765855490", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T06:25:01.230468+03:00", "updated_at": "2025-12-16T06:25:02.615822+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=1843b005877844c3867a9020ec15e636,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["543"],"Content-Type":["application/json"],"Sentry-Trace":["1843b005877844c3867a9020ec15e636-a3f5c54a5f298745"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 06:25:03.622867+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	Q2N8JLY	t	\N	127.0.0.1:41112	_	\N	f	production	\N	\N	200	\N
42	2025-12-16 06:25:11.788545+03	2025-12-16 06:25:11.788545+03	\N	intasend		payment	processed	{"invoice_id": "Q2N8JLY", "state": "COMPLETE", "provider": "CARD-PAYMENT", "charges": "0.09", "net_amount": "2.39", "currency": "KES", "value": "2.48", "account": "kamausimon217@gmail.com", "api_ref": "ORD-1-1765855490", "clearing_status": "CLEARING", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": "MASTERCARD"}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T06:25:01.230468+03:00", "updated_at": "2025-12-16T06:25:10.648588+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=d149e246255d445d8a3b166179a865ad,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["580"],"Content-Type":["application/json"],"Sentry-Trace":["d149e246255d445d8a3b166179a865ad-8a27dc1b4276ec85"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 06:25:11.752602+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	Q2N8JLY	t	\N	127.0.0.1:47152	_	\N	f	production	\N	\N	200	\N
43	2025-12-16 06:30:56.767125+03	2025-12-16 06:30:56.767125+03	\N	intasend		payment	processed	{"invoice_id": "R908MZR", "state": "PENDING", "provider": "CARD-PAYMENT", "charges": "0.00", "net_amount": "1.19", "currency": "KES", "value": "1.19", "account": "kamausimon217@gmail.com", "api_ref": "ORD-3-1765855843", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T06:30:55.165986+03:00", "updated_at": "2025-12-16T06:30:55.182248+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=6e96d404bdc8490a9ae141ee253cc8b2,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["540"],"Content-Type":["application/json"],"Sentry-Trace":["6e96d404bdc8490a9ae141ee253cc8b2-840d3aa2380d4b9c"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 06:30:56.608606+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	R908MZR	t	\N	127.0.0.1:51964	_	\N	f	production	\N	\N	200	\N
44	2025-12-16 06:30:57.681714+03	2025-12-16 06:30:57.681714+03	\N	intasend		payment	processed	{"invoice_id": "R908MZR", "state": "PROCESSING", "provider": "CARD-PAYMENT", "charges": "0.00", "net_amount": "1.19", "currency": "KES", "value": "1.19", "account": "kamausimon217@gmail.com", "api_ref": "ORD-3-1765855843", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T06:30:55.165986+03:00", "updated_at": "2025-12-16T06:30:56.366184+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=6e96d404bdc8490a9ae141ee253cc8b2,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["543"],"Content-Type":["application/json"],"Sentry-Trace":["6e96d404bdc8490a9ae141ee253cc8b2-99c2c562d92bf7c6"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 06:30:57.633297+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	R908MZR	t	\N	127.0.0.1:51980	_	\N	f	production	\N	\N	200	\N
45	2025-12-16 06:31:05.855648+03	2025-12-16 06:31:05.855648+03	\N	intasend		payment	processed	{"invoice_id": "R908MZR", "state": "COMPLETE", "provider": "CARD-PAYMENT", "charges": "0.05", "net_amount": "1.13", "currency": "KES", "value": "1.19", "account": "kamausimon217@gmail.com", "api_ref": "ORD-3-1765855843", "clearing_status": "CLEARING", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": "MASTERCARD"}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T06:30:55.165986+03:00", "updated_at": "2025-12-16T06:31:04.598965+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=a308ada3573645cabe611002aff7825a,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["580"],"Content-Type":["application/json"],"Sentry-Trace":["a308ada3573645cabe611002aff7825a-8932c324e965ce59"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 06:31:05.835398+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	R908MZR	t	\N	127.0.0.1:52554	_	\N	f	production	\N	\N	200	\N
46	2025-12-16 06:40:20.588364+03	2025-12-16 06:40:20.588364+03	\N	intasend		payment	processed	{"invoice_id": "Q628PKR", "state": "PENDING", "provider": "CARD-PAYMENT", "charges": "0.00", "net_amount": "2.38", "currency": "KES", "value": "2.38", "account": "kamausimon217@gmail.com", "api_ref": "ORD-4-1765856400", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T06:40:19.467605+03:00", "updated_at": "2025-12-16T06:40:19.480834+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=7957573aa24e48e7b06985b2cd0759d5,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["540"],"Content-Type":["application/json"],"Sentry-Trace":["7957573aa24e48e7b06985b2cd0759d5-ad3dfcfafc672c19"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 06:40:20.56512+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	Q628PKR	t	\N	127.0.0.1:58870	_	\N	f	production	\N	\N	200	\N
47	2025-12-16 06:40:21.678196+03	2025-12-16 06:40:21.678196+03	\N	intasend		payment	processed	{"invoice_id": "Q628PKR", "state": "PROCESSING", "provider": "CARD-PAYMENT", "charges": "0.00", "net_amount": "2.38", "currency": "KES", "value": "2.38", "account": "kamausimon217@gmail.com", "api_ref": "ORD-4-1765856400", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T06:40:19.467605+03:00", "updated_at": "2025-12-16T06:40:20.611518+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=7957573aa24e48e7b06985b2cd0759d5,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["543"],"Content-Type":["application/json"],"Sentry-Trace":["7957573aa24e48e7b06985b2cd0759d5-b02c63992bb1cdd5"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 06:40:21.65752+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	Q628PKR	t	\N	127.0.0.1:47274	_	\N	f	production	\N	\N	200	\N
48	2025-12-16 06:40:30.745012+03	2025-12-16 06:40:30.745012+03	\N	intasend		payment	processed	{"invoice_id": "Q628PKR", "state": "COMPLETE", "provider": "CARD-PAYMENT", "charges": "0.09", "net_amount": "2.29", "currency": "KES", "value": "2.38", "account": "kamausimon217@gmail.com", "api_ref": "ORD-4-1765856400", "clearing_status": "CLEARING", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": "MASTERCARD"}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T06:40:19.467605+03:00", "updated_at": "2025-12-16T06:40:29.587967+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=5dbc906b7a134355878ab00bfeef6318,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["580"],"Content-Type":["application/json"],"Sentry-Trace":["5dbc906b7a134355878ab00bfeef6318-8be12eff82a30211"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 06:40:30.680479+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	Q628PKR	t	\N	127.0.0.1:47284	_	\N	f	production	\N	\N	200	\N
49	2025-12-16 06:52:39.714621+03	2025-12-16 06:52:39.714621+03	\N	intasend		payment	processed	{"invoice_id": "QJOD5DR", "state": "PENDING", "provider": "CARD-PAYMENT", "charges": "0.00", "net_amount": "2.38", "currency": "KES", "value": "2.38", "account": "kamausimon217@gmail.com", "api_ref": "ORD-5-1765857147", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T06:52:38.571650+03:00", "updated_at": "2025-12-16T06:52:38.589513+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=f2d1917df82d4335811d48412d873d28,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["540"],"Content-Type":["application/json"],"Sentry-Trace":["f2d1917df82d4335811d48412d873d28-aff24b9acd29fd56"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 06:52:39.693446+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	QJOD5DR	t	\N	127.0.0.1:36858	_	\N	f	production	\N	\N	200	\N
50	2025-12-16 06:52:40.905215+03	2025-12-16 06:52:40.905215+03	\N	intasend		payment	processed	{"invoice_id": "QJOD5DR", "state": "PROCESSING", "provider": "CARD-PAYMENT", "charges": "0.00", "net_amount": "2.38", "currency": "KES", "value": "2.38", "account": "kamausimon217@gmail.com", "api_ref": "ORD-5-1765857147", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T06:52:38.571650+03:00", "updated_at": "2025-12-16T06:52:39.811216+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=f2d1917df82d4335811d48412d873d28,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["543"],"Content-Type":["application/json"],"Sentry-Trace":["f2d1917df82d4335811d48412d873d28-b00d23beda5aeffd"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 06:52:40.881761+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	QJOD5DR	t	\N	127.0.0.1:36870	_	\N	f	production	\N	\N	200	\N
51	2025-12-16 06:52:49.609715+03	2025-12-16 06:52:49.609715+03	\N	intasend		payment	processed	{"invoice_id": "QJOD5DR", "state": "COMPLETE", "provider": "CARD-PAYMENT", "charges": "0.09", "net_amount": "2.29", "currency": "KES", "value": "2.38", "account": "kamausimon217@gmail.com", "api_ref": "ORD-5-1765857147", "clearing_status": "CLEARING", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": "MASTERCARD"}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T06:52:38.571650+03:00", "updated_at": "2025-12-16T06:52:48.569697+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=17e5a8377fb945d087385f83e911e488,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["580"],"Content-Type":["application/json"],"Sentry-Trace":["17e5a8377fb945d087385f83e911e488-8f42005e8c20d9f2"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 06:52:49.583008+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	QJOD5DR	t	\N	127.0.0.1:55906	_	\N	f	production	\N	\N	200	\N
52	2025-12-16 06:55:38.708283+03	2025-12-16 06:55:38.708283+03	\N	intasend		payment	processed	{"invoice_id": "RKP7OEY", "state": "PENDING", "provider": "CARD-PAYMENT", "charges": "0.00", "net_amount": "2.38", "currency": "KES", "value": "2.38", "account": "kamausimon217@gmail.com", "api_ref": "ORD-6-1765857327", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T06:55:37.551704+03:00", "updated_at": "2025-12-16T06:55:37.566054+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=cdb220b50d024947b75ea3a6e446a593,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["540"],"Content-Type":["application/json"],"Sentry-Trace":["cdb220b50d024947b75ea3a6e446a593-8a2bc6c82083b97f"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 06:55:38.67868+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	RKP7OEY	t	\N	127.0.0.1:52062	_	\N	f	production	\N	\N	200	\N
53	2025-12-16 06:55:39.826948+03	2025-12-16 06:55:39.826948+03	\N	intasend		payment	processed	{"invoice_id": "RKP7OEY", "state": "PROCESSING", "provider": "CARD-PAYMENT", "charges": "0.00", "net_amount": "2.38", "currency": "KES", "value": "2.38", "account": "kamausimon217@gmail.com", "api_ref": "ORD-6-1765857327", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T06:55:37.551704+03:00", "updated_at": "2025-12-16T06:55:38.724497+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=cdb220b50d024947b75ea3a6e446a593,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["543"],"Content-Type":["application/json"],"Sentry-Trace":["cdb220b50d024947b75ea3a6e446a593-8069a53f4b1b37c3"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 06:55:39.767193+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	RKP7OEY	t	\N	127.0.0.1:52064	_	\N	f	production	\N	\N	200	\N
54	2025-12-16 06:55:47.742961+03	2025-12-16 06:55:47.742961+03	\N	intasend		payment	processed	{"invoice_id": "RKP7OEY", "state": "COMPLETE", "provider": "CARD-PAYMENT", "charges": "0.09", "net_amount": "2.29", "currency": "KES", "value": "2.38", "account": "kamausimon217@gmail.com", "api_ref": "ORD-6-1765857327", "clearing_status": "CLEARING", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": "MASTERCARD"}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T06:55:37.551704+03:00", "updated_at": "2025-12-16T06:55:46.636546+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=854a470ae93240fdaecd51a98a7d546c,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["580"],"Content-Type":["application/json"],"Sentry-Trace":["854a470ae93240fdaecd51a98a7d546c-92432da354d63fce"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 06:55:47.71318+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	RKP7OEY	t	\N	127.0.0.1:60954	_	\N	f	production	\N	\N	200	\N
55	2025-12-16 07:08:30.660937+03	2025-12-16 07:08:30.660937+03	\N	intasend		payment	processed	{"invoice_id": "R8PG3VR", "state": "PENDING", "provider": "CARD-PAYMENT", "charges": "0.00", "net_amount": "2.38", "currency": "KES", "value": "2.38", "account": "kamausimon217@gmail.com", "api_ref": "ORD-7-1765858111", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T07:08:29.533584+03:00", "updated_at": "2025-12-16T07:08:29.549713+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=dbead74afe034af995dec44eca495bfb,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["540"],"Content-Type":["application/json"],"Sentry-Trace":["dbead74afe034af995dec44eca495bfb-b23582f810865970"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 07:08:30.629815+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	R8PG3VR	t	\N	127.0.0.1:40176	_	\N	f	production	\N	\N	200	\N
56	2025-12-16 07:08:31.63169+03	2025-12-16 07:08:31.63169+03	\N	intasend		payment	processed	{"invoice_id": "R8PG3VR", "state": "PROCESSING", "provider": "CARD-PAYMENT", "charges": "0.00", "net_amount": "2.38", "currency": "KES", "value": "2.38", "account": "kamausimon217@gmail.com", "api_ref": "ORD-7-1765858111", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T07:08:29.533584+03:00", "updated_at": "2025-12-16T07:08:30.669202+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=dbead74afe034af995dec44eca495bfb,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["543"],"Content-Type":["application/json"],"Sentry-Trace":["dbead74afe034af995dec44eca495bfb-9173d5002b9fbe1a"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 07:08:31.609068+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	R8PG3VR	t	\N	127.0.0.1:42886	_	\N	f	production	\N	\N	200	\N
57	2025-12-16 07:08:39.898772+03	2025-12-16 07:08:39.898772+03	\N	intasend		payment	processed	{"invoice_id": "R8PG3VR", "state": "COMPLETE", "provider": "CARD-PAYMENT", "charges": "0.09", "net_amount": "2.29", "currency": "KES", "value": "2.38", "account": "kamausimon217@gmail.com", "api_ref": "ORD-7-1765858111", "clearing_status": "CLEARING", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": "MASTERCARD"}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T07:08:29.533584+03:00", "updated_at": "2025-12-16T07:08:38.647333+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=0ff5192bb03e4301bbccc6e7d07ea6e8,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["580"],"Content-Type":["application/json"],"Sentry-Trace":["0ff5192bb03e4301bbccc6e7d07ea6e8-9af57691f5539e62"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 07:08:39.878143+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	R8PG3VR	t	\N	127.0.0.1:42902	_	\N	f	production	\N	\N	200	\N
58	2025-12-16 07:14:20.745855+03	2025-12-16 07:14:20.745855+03	\N	intasend		payment	processed	{"invoice_id": "QP6E2ZY", "state": "PENDING", "provider": "CARD-PAYMENT", "charges": "0.00", "net_amount": "1.19", "currency": "KES", "value": "1.19", "account": "kamausimon217@gmail.com", "api_ref": "ORD-7-1765858465", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T07:14:19.611924+03:00", "updated_at": "2025-12-16T07:14:19.630346+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=dbda98623d224941944c5efd1ab2b3df,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["540"],"Content-Type":["application/json"],"Sentry-Trace":["dbda98623d224941944c5efd1ab2b3df-bea4a8b48e4fa118"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 07:14:20.697822+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	QP6E2ZY	t	\N	127.0.0.1:58424	_	\N	f	production	\N	\N	200	\N
59	2025-12-16 07:14:21.841369+03	2025-12-16 07:14:21.841369+03	\N	intasend		payment	processed	{"invoice_id": "QP6E2ZY", "state": "PROCESSING", "provider": "CARD-PAYMENT", "charges": "0.00", "net_amount": "1.19", "currency": "KES", "value": "1.19", "account": "kamausimon217@gmail.com", "api_ref": "ORD-7-1765858465", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T07:14:19.611924+03:00", "updated_at": "2025-12-16T07:14:20.808198+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=dbda98623d224941944c5efd1ab2b3df,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["543"],"Content-Type":["application/json"],"Sentry-Trace":["dbda98623d224941944c5efd1ab2b3df-8c1f7ddf87851fad"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 07:14:21.818471+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	QP6E2ZY	t	\N	127.0.0.1:50354	_	\N	f	production	\N	\N	200	\N
60	2025-12-16 07:14:29.792223+03	2025-12-16 07:14:29.792223+03	\N	intasend		payment	processed	{"invoice_id": "QP6E2ZY", "state": "COMPLETE", "provider": "CARD-PAYMENT", "charges": "0.05", "net_amount": "1.13", "currency": "KES", "value": "1.19", "account": "kamausimon217@gmail.com", "api_ref": "ORD-7-1765858465", "clearing_status": "CLEARING", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": "MASTERCARD"}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T07:14:19.611924+03:00", "updated_at": "2025-12-16T07:14:28.616971+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=467f7486bb9d401a9ca748c2fda49267,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["580"],"Content-Type":["application/json"],"Sentry-Trace":["467f7486bb9d401a9ca748c2fda49267-929893762ac8000c"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 07:14:29.747615+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	QP6E2ZY	t	\N	127.0.0.1:50362	_	\N	f	production	\N	\N	200	\N
61	2025-12-16 07:24:23.114478+03	2025-12-16 07:24:23.190133+03	\N	intasend		payment	processed	{"invoice_id": "R0Z8KGR", "state": "PENDING", "provider": "CARD-PAYMENT", "charges": "0.00", "net_amount": "1.19", "currency": "KES", "value": "1.19", "account": "kamausimon217@gmail.com", "api_ref": "ORD-9-1765859047", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T07:24:06.814833+03:00", "updated_at": "2025-12-16T07:24:06.832766+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=1bf6ad11ba4748c6873ca99ada7aa0d1,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["540"],"Content-Type":["application/json"],"Sentry-Trace":["1bf6ad11ba4748c6873ca99ada7aa0d1-8a17bcfe1703f640"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 07:24:23.089309+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	R0Z8KGR	t	\N	127.0.0.1:42658	_	\N	f	production	\N	\N	200	\N
62	2025-12-16 07:24:24.154702+03	2025-12-16 07:24:24.31961+03	\N	intasend		payment	processed	{"invoice_id": "R0Z8KGR", "state": "PROCESSING", "provider": "CARD-PAYMENT", "charges": "0.00", "net_amount": "1.19", "currency": "KES", "value": "1.19", "account": "kamausimon217@gmail.com", "api_ref": "ORD-9-1765859047", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": null}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T07:24:06.814833+03:00", "updated_at": "2025-12-16T07:24:08.046213+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=1bf6ad11ba4748c6873ca99ada7aa0d1,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["543"],"Content-Type":["application/json"],"Sentry-Trace":["1bf6ad11ba4748c6873ca99ada7aa0d1-9a4ffa173b56c02d"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 07:24:24.117978+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	R0Z8KGR	t	\N	127.0.0.1:42672	_	\N	f	production	\N	\N	200	\N
63	2025-12-16 07:24:17.696927+03	2025-12-16 07:24:18.048585+03	\N	intasend		payment	processed	{"invoice_id": "R0Z8KGR", "state": "COMPLETE", "provider": "CARD-PAYMENT", "charges": "0.05", "net_amount": "1.13", "currency": "KES", "value": "1.19", "account": "kamausimon217@gmail.com", "api_ref": "ORD-9-1765859047", "clearing_status": "CLEARING", "mpesa_reference": null, "host": "127.0.0.1", "card_info": {"bin_country": null, "card_type": "MASTERCARD"}, "retry_count": 0, "failed_reason": null, "failed_code": null, "failed_code_link": null, "created_at": "2025-12-16T07:24:06.814833+03:00", "updated_at": "2025-12-16T07:24:16.730709+03:00", "challenge": "L@Eh5sMbXGC5t7C"}	{"Accept":["*/*"],"Accept-Encoding":["identity"],"Baggage":["sentry-trace_id=7441bd47bf8046459fe4cc42922919ff,sentry-environment=sandbox,sentry-release=f55f3a5a1781a42ff199f15dffb9574c55988967,sentry-public_key=2e5a2191c743449384500d09e0fef9fd"],"Cache-Control":["no-cache"],"Content-Length":["580"],"Content-Type":["application/json"],"Sentry-Trace":["7441bd47bf8046459fe4cc42922919ff-b66fdeb4a8ca6190"],"User-Agent":["_"],"X-Forwarded-For":["157.245.201.212"],"X-Forwarded-Host":["ticketingapp.ngrok.dev"],"X-Forwarded-Proto":["https","https"],"X-Forwarded-Ssl":["on"]}	POST		2025-12-16 07:24:17.675829+03	\N	0	\N	t	\N	\N	\N	\N	\N	\N	\N	\N	R0Z8KGR	t	\N	127.0.0.1:48456	_	\N	f	production	\N	\N	200	\N
\.


--
-- Name: account_activities_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.account_activities_id_seq', 185, true);


--
-- Name: account_payment_gateways_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.account_payment_gateways_id_seq', 1, false);


--
-- Name: accounts_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.accounts_id_seq', 11, true);


--
-- Name: attendees_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.attendees_id_seq', 1, false);


--
-- Name: currencies_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.currencies_id_seq', 14, true);


--
-- Name: date_formats_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.date_formats_id_seq', 8, true);


--
-- Name: date_time_formats_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.date_time_formats_id_seq', 8, true);


--
-- Name: email_verifications_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.email_verifications_id_seq', 10, true);


--
-- Name: event_images_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.event_images_id_seq', 11, true);


--
-- Name: event_metrics_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.event_metrics_id_seq', 1, false);


--
-- Name: event_stats_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.event_stats_id_seq', 1, false);


--
-- Name: event_venues_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.event_venues_id_seq', 1, false);


--
-- Name: events_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.events_id_seq', 4, true);


--
-- Name: login_history_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.login_history_id_seq', 139, true);


--
-- Name: notification_preferences_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.notification_preferences_id_seq', 1, true);


--
-- Name: order_items_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.order_items_id_seq', 9, true);


--
-- Name: orders_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.orders_id_seq', 9, true);


--
-- Name: organizers_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.organizers_id_seq', 1, true);


--
-- Name: password_reset_attempts_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.password_reset_attempts_id_seq', 4, true);


--
-- Name: password_resets_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.password_resets_id_seq', 3, true);


--
-- Name: payment_gateways_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.payment_gateways_id_seq', 1, false);


--
-- Name: payment_methods_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.payment_methods_id_seq', 1, false);


--
-- Name: payment_records_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.payment_records_id_seq', 29, true);


--
-- Name: payment_transactions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.payment_transactions_id_seq', 10, true);


--
-- Name: payout_accounts_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.payout_accounts_id_seq', 1, false);


--
-- Name: promotion_rules_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.promotion_rules_id_seq', 1, false);


--
-- Name: promotion_usages_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.promotion_usages_id_seq', 1, false);


--
-- Name: promotions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.promotions_id_seq', 6, true);


--
-- Name: recovery_codes_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.recovery_codes_id_seq', 10, true);


--
-- Name: refund_line_items_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.refund_line_items_id_seq', 1, false);


--
-- Name: refund_records_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.refund_records_id_seq', 1, false);


--
-- Name: reserved_tickets_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.reserved_tickets_id_seq', 4, true);


--
-- Name: reset_configurations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.reset_configurations_id_seq', 2, true);


--
-- Name: security_metrics_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.security_metrics_id_seq', 1, false);


--
-- Name: settlement_items_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.settlement_items_id_seq', 1, false);


--
-- Name: settlement_records_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.settlement_records_id_seq', 1, false);


--
-- Name: support_ticket_comments_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.support_ticket_comments_id_seq', 2, true);


--
-- Name: support_tickets_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.support_tickets_id_seq', 3, true);


--
-- Name: system_metrics_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.system_metrics_id_seq', 1, false);


--
-- Name: ticket_classes_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.ticket_classes_id_seq', 5, true);


--
-- Name: ticket_orders_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.ticket_orders_id_seq', 1, false);


--
-- Name: ticket_transfer_histories_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.ticket_transfer_histories_id_seq', 2, true);


--
-- Name: tickets_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.tickets_id_seq', 12, true);


--
-- Name: timezones_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.timezones_id_seq', 16, true);


--
-- Name: two_factor_attempts_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.two_factor_attempts_id_seq', 6, true);


--
-- Name: two_factor_auths_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.two_factor_auths_id_seq', 6, true);


--
-- Name: two_factor_sessions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.two_factor_sessions_id_seq', 4, true);


--
-- Name: user_engagement_metrics_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.user_engagement_metrics_id_seq', 1, false);


--
-- Name: users_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.users_id_seq', 10, true);


--
-- Name: venues_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.venues_id_seq', 1, false);


--
-- Name: waitlist_entries_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.waitlist_entries_id_seq', 1, true);


--
-- Name: webhook_logs_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.webhook_logs_id_seq', 63, true);


--
-- Name: account_activities account_activities_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.account_activities
    ADD CONSTRAINT account_activities_pkey PRIMARY KEY (id);


--
-- Name: account_payment_gateways account_payment_gateways_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.account_payment_gateways
    ADD CONSTRAINT account_payment_gateways_pkey PRIMARY KEY (id);


--
-- Name: accounts accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.accounts
    ADD CONSTRAINT accounts_pkey PRIMARY KEY (id);


--
-- Name: attendees attendees_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.attendees
    ADD CONSTRAINT attendees_pkey PRIMARY KEY (id);


--
-- Name: currencies currencies_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.currencies
    ADD CONSTRAINT currencies_pkey PRIMARY KEY (id);


--
-- Name: date_formats date_formats_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.date_formats
    ADD CONSTRAINT date_formats_pkey PRIMARY KEY (id);


--
-- Name: date_time_formats date_time_formats_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.date_time_formats
    ADD CONSTRAINT date_time_formats_pkey PRIMARY KEY (id);


--
-- Name: email_verifications email_verifications_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.email_verifications
    ADD CONSTRAINT email_verifications_pkey PRIMARY KEY (id);


--
-- Name: event_images event_images_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.event_images
    ADD CONSTRAINT event_images_pkey PRIMARY KEY (id);


--
-- Name: event_metrics event_metrics_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.event_metrics
    ADD CONSTRAINT event_metrics_pkey PRIMARY KEY (id);


--
-- Name: event_stats event_stats_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.event_stats
    ADD CONSTRAINT event_stats_pkey PRIMARY KEY (id);


--
-- Name: event_venues event_venues_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.event_venues
    ADD CONSTRAINT event_venues_pkey PRIMARY KEY (id);


--
-- Name: events events_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.events
    ADD CONSTRAINT events_pkey PRIMARY KEY (id);


--
-- Name: login_history login_history_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.login_history
    ADD CONSTRAINT login_history_pkey PRIMARY KEY (id);


--
-- Name: notification_preferences notification_preferences_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notification_preferences
    ADD CONSTRAINT notification_preferences_pkey PRIMARY KEY (id);


--
-- Name: order_items order_items_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_pkey PRIMARY KEY (id);


--
-- Name: orders orders_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);


--
-- Name: organizers organizers_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.organizers
    ADD CONSTRAINT organizers_pkey PRIMARY KEY (id);


--
-- Name: password_reset_attempts password_reset_attempts_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.password_reset_attempts
    ADD CONSTRAINT password_reset_attempts_pkey PRIMARY KEY (id);


--
-- Name: password_resets password_resets_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.password_resets
    ADD CONSTRAINT password_resets_pkey PRIMARY KEY (id);


--
-- Name: payment_gateways payment_gateways_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payment_gateways
    ADD CONSTRAINT payment_gateways_pkey PRIMARY KEY (id);


--
-- Name: payment_methods payment_methods_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payment_methods
    ADD CONSTRAINT payment_methods_pkey PRIMARY KEY (id);


--
-- Name: payment_records payment_records_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payment_records
    ADD CONSTRAINT payment_records_pkey PRIMARY KEY (id);


--
-- Name: payment_transactions payment_transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payment_transactions
    ADD CONSTRAINT payment_transactions_pkey PRIMARY KEY (id);


--
-- Name: payout_accounts payout_accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payout_accounts
    ADD CONSTRAINT payout_accounts_pkey PRIMARY KEY (id);


--
-- Name: promotion_rules promotion_rules_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.promotion_rules
    ADD CONSTRAINT promotion_rules_pkey PRIMARY KEY (id);


--
-- Name: promotion_usages promotion_usages_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.promotion_usages
    ADD CONSTRAINT promotion_usages_pkey PRIMARY KEY (id);


--
-- Name: promotions promotions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.promotions
    ADD CONSTRAINT promotions_pkey PRIMARY KEY (id);


--
-- Name: recovery_codes recovery_codes_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.recovery_codes
    ADD CONSTRAINT recovery_codes_pkey PRIMARY KEY (id);


--
-- Name: refund_line_items refund_line_items_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.refund_line_items
    ADD CONSTRAINT refund_line_items_pkey PRIMARY KEY (id);


--
-- Name: refund_records refund_records_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.refund_records
    ADD CONSTRAINT refund_records_pkey PRIMARY KEY (id);


--
-- Name: reserved_tickets reserved_tickets_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.reserved_tickets
    ADD CONSTRAINT reserved_tickets_pkey PRIMARY KEY (id);


--
-- Name: reset_configurations reset_configurations_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.reset_configurations
    ADD CONSTRAINT reset_configurations_pkey PRIMARY KEY (id);


--
-- Name: security_metrics security_metrics_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.security_metrics
    ADD CONSTRAINT security_metrics_pkey PRIMARY KEY (id);


--
-- Name: settlement_items settlement_items_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.settlement_items
    ADD CONSTRAINT settlement_items_pkey PRIMARY KEY (id);


--
-- Name: settlement_payment_records settlement_payment_records_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.settlement_payment_records
    ADD CONSTRAINT settlement_payment_records_pkey PRIMARY KEY (settlement_item_id, payment_record_id);


--
-- Name: settlement_records settlement_records_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.settlement_records
    ADD CONSTRAINT settlement_records_pkey PRIMARY KEY (id);


--
-- Name: support_ticket_comments support_ticket_comments_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.support_ticket_comments
    ADD CONSTRAINT support_ticket_comments_pkey PRIMARY KEY (id);


--
-- Name: support_tickets support_tickets_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.support_tickets
    ADD CONSTRAINT support_tickets_pkey PRIMARY KEY (id);


--
-- Name: system_metrics system_metrics_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.system_metrics
    ADD CONSTRAINT system_metrics_pkey PRIMARY KEY (id);


--
-- Name: ticket_classes ticket_classes_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ticket_classes
    ADD CONSTRAINT ticket_classes_pkey PRIMARY KEY (id);


--
-- Name: ticket_orders ticket_orders_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ticket_orders
    ADD CONSTRAINT ticket_orders_pkey PRIMARY KEY (id);


--
-- Name: ticket_transfer_histories ticket_transfer_histories_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ticket_transfer_histories
    ADD CONSTRAINT ticket_transfer_histories_pkey PRIMARY KEY (id);


--
-- Name: tickets tickets_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tickets
    ADD CONSTRAINT tickets_pkey PRIMARY KEY (id);


--
-- Name: timezones timezones_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.timezones
    ADD CONSTRAINT timezones_pkey PRIMARY KEY (id);


--
-- Name: two_factor_attempts two_factor_attempts_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.two_factor_attempts
    ADD CONSTRAINT two_factor_attempts_pkey PRIMARY KEY (id);


--
-- Name: two_factor_auths two_factor_auths_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.two_factor_auths
    ADD CONSTRAINT two_factor_auths_pkey PRIMARY KEY (id);


--
-- Name: two_factor_sessions two_factor_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.two_factor_sessions
    ADD CONSTRAINT two_factor_sessions_pkey PRIMARY KEY (id);


--
-- Name: password_resets uni_password_resets_token; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.password_resets
    ADD CONSTRAINT uni_password_resets_token UNIQUE (token);


--
-- Name: promotions uni_promotions_code; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.promotions
    ADD CONSTRAINT uni_promotions_code UNIQUE (code);


--
-- Name: refund_records uni_refund_records_refund_number; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.refund_records
    ADD CONSTRAINT uni_refund_records_refund_number UNIQUE (refund_number);


--
-- Name: reset_configurations uni_reset_configurations_config_name; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.reset_configurations
    ADD CONSTRAINT uni_reset_configurations_config_name UNIQUE (config_name);


--
-- Name: settlement_records uni_settlement_records_settlement_batch_id; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.settlement_records
    ADD CONSTRAINT uni_settlement_records_settlement_batch_id UNIQUE (settlement_batch_id);


--
-- Name: support_tickets uni_support_tickets_ticket_number; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.support_tickets
    ADD CONSTRAINT uni_support_tickets_ticket_number UNIQUE (ticket_number);


--
-- Name: tickets uni_tickets_qr_code; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tickets
    ADD CONSTRAINT uni_tickets_qr_code UNIQUE (qr_code);


--
-- Name: tickets uni_tickets_ticket_number; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tickets
    ADD CONSTRAINT uni_tickets_ticket_number UNIQUE (ticket_number);


--
-- Name: user_engagement_metrics user_engagement_metrics_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_engagement_metrics
    ADD CONSTRAINT user_engagement_metrics_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: venues venues_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.venues
    ADD CONSTRAINT venues_pkey PRIMARY KEY (id);


--
-- Name: waitlist_entries waitlist_entries_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.waitlist_entries
    ADD CONSTRAINT waitlist_entries_pkey PRIMARY KEY (id);


--
-- Name: webhook_logs webhook_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.webhook_logs
    ADD CONSTRAINT webhook_logs_pkey PRIMARY KEY (id);


--
-- Name: idx_2fa_attempts_recent; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_2fa_attempts_recent ON public.two_factor_attempts USING btree (user_id, attempted_at, success);


--
-- Name: idx_2fa_enabled_users; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_2fa_enabled_users ON public.two_factor_auths USING btree (user_id, enabled) WHERE (enabled = true);


--
-- Name: idx_2fa_sessions_active; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_2fa_sessions_active ON public.two_factor_sessions USING btree (user_id, expires_at, verified) WHERE (verified = false);


--
-- Name: idx_account_activities_category; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_account_activities_category ON public.account_activities USING btree (category);


--
-- Name: idx_account_activities_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_account_activities_deleted_at ON public.account_activities USING btree (deleted_at);


--
-- Name: idx_account_activities_success; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_account_activities_success ON public.account_activities USING btree (success);


--
-- Name: idx_account_activities_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_account_activities_user_id ON public.account_activities USING btree (user_id);


--
-- Name: idx_account_payment_gateways_account_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_account_payment_gateways_account_id ON public.account_payment_gateways USING btree (account_id);


--
-- Name: idx_account_payment_gateways_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_account_payment_gateways_deleted_at ON public.account_payment_gateways USING btree (deleted_at);


--
-- Name: idx_account_payment_gateways_payment_gateway_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_account_payment_gateways_payment_gateway_id ON public.account_payment_gateways USING btree (payment_gateway_id);


--
-- Name: idx_accounts_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_accounts_deleted_at ON public.accounts USING btree (deleted_at);


--
-- Name: idx_accounts_email; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_accounts_email ON public.accounts USING btree (email);


--
-- Name: idx_activities_by_action; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_activities_by_action ON public.account_activities USING btree (action, success, "timestamp" DESC);


--
-- Name: idx_activities_by_category; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_activities_by_category ON public.account_activities USING btree (category, "timestamp" DESC);


--
-- Name: idx_activities_failed; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_activities_failed ON public.account_activities USING btree (success, severity, "timestamp" DESC) WHERE (success = false);


--
-- Name: idx_activities_recent; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_activities_recent ON public.account_activities USING btree (account_id, "timestamp" DESC);


--
-- Name: idx_activity_account_time; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_activity_account_time ON public.account_activities USING btree (account_id, "timestamp");


--
-- Name: idx_activity_action; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_activity_action ON public.account_activities USING btree (action);


--
-- Name: idx_activity_timestamp; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_activity_timestamp ON public.account_activities USING btree ("timestamp");


--
-- Name: idx_attendees_account_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_attendees_account_id ON public.attendees USING btree (account_id);


--
-- Name: idx_attendees_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_attendees_deleted_at ON public.attendees USING btree (deleted_at);


--
-- Name: idx_attendees_event_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_attendees_event_id ON public.attendees USING btree (event_id);


--
-- Name: idx_attendees_order_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_attendees_order_id ON public.attendees USING btree (order_id);


--
-- Name: idx_attendees_ticket_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_attendees_ticket_id ON public.attendees USING btree (ticket_id);


--
-- Name: idx_currencies_code; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_currencies_code ON public.currencies USING btree (code);


--
-- Name: idx_currencies_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_currencies_deleted_at ON public.currencies USING btree (deleted_at);


--
-- Name: idx_date_formats_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_date_formats_deleted_at ON public.date_formats USING btree (deleted_at);


--
-- Name: idx_date_formats_format; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_date_formats_format ON public.date_formats USING btree (format);


--
-- Name: idx_date_time_formats_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_date_time_formats_deleted_at ON public.date_time_formats USING btree (deleted_at);


--
-- Name: idx_date_time_formats_format; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_date_time_formats_format ON public.date_time_formats USING btree (format);


--
-- Name: idx_email_verification_email; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_email_verification_email ON public.email_verifications USING btree (email);


--
-- Name: idx_email_verification_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_email_verification_status ON public.email_verifications USING btree (status);


--
-- Name: idx_email_verification_user; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_email_verification_user ON public.email_verifications USING btree (user_id);


--
-- Name: idx_email_verifications_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_email_verifications_deleted_at ON public.email_verifications USING btree (deleted_at);


--
-- Name: idx_email_verifications_expires_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_email_verifications_expires_at ON public.email_verifications USING btree (expires_at);


--
-- Name: idx_email_verifications_token; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_email_verifications_token ON public.email_verifications USING btree (token);


--
-- Name: idx_event_analytics; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_event_analytics ON public.event_metrics USING btree (event_id, date);


--
-- Name: idx_event_images_account_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_event_images_account_id ON public.event_images USING btree (account_id);


--
-- Name: idx_event_images_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_event_images_deleted_at ON public.event_images USING btree (deleted_at);


--
-- Name: idx_event_images_event_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_event_images_event_id ON public.event_images USING btree (event_id);


--
-- Name: idx_event_images_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_event_images_user_id ON public.event_images USING btree (user_id);


--
-- Name: idx_event_metrics; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_event_metrics ON public.event_metrics USING btree (event_id, date);


--
-- Name: idx_event_metrics_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_event_metrics_deleted_at ON public.event_metrics USING btree (deleted_at);


--
-- Name: idx_event_stats_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_event_stats_deleted_at ON public.event_stats USING btree (deleted_at);


--
-- Name: idx_event_venue; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_event_venue ON public.event_venues USING btree (venue_id, event_id);


--
-- Name: idx_event_venues_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_event_venues_deleted_at ON public.event_venues USING btree (deleted_at);


--
-- Name: idx_events_account_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_events_account_id ON public.events USING btree (account_id);


--
-- Name: idx_events_category; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_events_category ON public.events USING btree (category);


--
-- Name: idx_events_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_events_deleted_at ON public.events USING btree (deleted_at);


--
-- Name: idx_events_organizer_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_events_organizer_id ON public.events USING btree (organizer_id);


--
-- Name: idx_events_stats_lookup; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_events_stats_lookup ON public.event_stats USING btree (date, event_id);


--
-- Name: idx_login_history_account_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_login_history_account_id ON public.login_history USING btree (account_id);


--
-- Name: idx_login_history_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_login_history_deleted_at ON public.login_history USING btree (deleted_at);


--
-- Name: idx_login_history_failed; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_login_history_failed ON public.login_history USING btree (ip_address, success, login_at DESC) WHERE (success = false);


--
-- Name: idx_login_history_ip_address; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_login_history_ip_address ON public.login_history USING btree (ip_address);


--
-- Name: idx_login_history_login_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_login_history_login_at ON public.login_history USING btree (login_at);


--
-- Name: idx_login_history_recent; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_login_history_recent ON public.login_history USING btree (account_id, login_at DESC);


--
-- Name: idx_login_history_success; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_login_history_success ON public.login_history USING btree (success);


--
-- Name: idx_login_history_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_login_history_user_id ON public.login_history USING btree (user_id);


--
-- Name: idx_metric_lookup; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_metric_lookup ON public.system_metrics USING btree (metric_name, metric_type, granularity, "timestamp");


--
-- Name: idx_metrics_time_series; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_metrics_time_series ON public.system_metrics USING btree (metric_name, granularity, "timestamp");


--
-- Name: idx_notification_preferences_account_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_notification_preferences_account_id ON public.notification_preferences USING btree (account_id);


--
-- Name: idx_notification_preferences_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_notification_preferences_deleted_at ON public.notification_preferences USING btree (deleted_at);


--
-- Name: idx_order_items_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_order_items_deleted_at ON public.order_items USING btree (deleted_at);


--
-- Name: idx_order_items_order_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_order_items_order_id ON public.order_items USING btree (order_id);


--
-- Name: idx_order_items_ticket_class_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_order_items_ticket_class_id ON public.order_items USING btree (ticket_class_id);


--
-- Name: idx_order_usage; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_order_usage ON public.promotion_usages USING btree (order_id);


--
-- Name: idx_orders_account_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_orders_account_id ON public.orders USING btree (account_id);


--
-- Name: idx_orders_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_orders_deleted_at ON public.orders USING btree (deleted_at);


--
-- Name: idx_orders_event_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_orders_event_id ON public.orders USING btree (event_id);


--
-- Name: idx_orders_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_orders_status ON public.orders USING btree (status);


--
-- Name: idx_organizers_account_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_organizers_account_id ON public.organizers USING btree (account_id);


--
-- Name: idx_organizers_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_organizers_deleted_at ON public.organizers USING btree (deleted_at);


--
-- Name: idx_organizers_payment_gateway_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_organizers_payment_gateway_id ON public.organizers USING btree (payment_gateway_id);


--
-- Name: idx_password_reset_attempts_attempted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_password_reset_attempts_attempted_at ON public.password_reset_attempts USING btree (attempted_at);


--
-- Name: idx_password_reset_attempts_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_password_reset_attempts_deleted_at ON public.password_reset_attempts USING btree (deleted_at);


--
-- Name: idx_password_reset_attempts_ip_address; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_password_reset_attempts_ip_address ON public.password_reset_attempts USING btree (ip_address);


--
-- Name: idx_password_reset_attempts_password_reset_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_password_reset_attempts_password_reset_id ON public.password_reset_attempts USING btree (password_reset_id);


--
-- Name: idx_password_resets_account_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_password_resets_account_id ON public.password_resets USING btree (account_id);


--
-- Name: idx_password_resets_approved_by; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_password_resets_approved_by ON public.password_resets USING btree (approved_by);


--
-- Name: idx_password_resets_cleanup; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_password_resets_cleanup ON public.password_resets USING btree (cleanup_after, should_cleanup) WHERE (should_cleanup = true);


--
-- Name: idx_password_resets_cleanup_after; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_password_resets_cleanup_after ON public.password_resets USING btree (cleanup_after);


--
-- Name: idx_password_resets_cooldown_until; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_password_resets_cooldown_until ON public.password_resets USING btree (cooldown_until);


--
-- Name: idx_password_resets_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_password_resets_deleted_at ON public.password_resets USING btree (deleted_at, deleted_at);


--
-- Name: idx_password_resets_email; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_password_resets_email ON public.password_resets USING btree (email);


--
-- Name: idx_password_resets_expires_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_password_resets_expires_at ON public.password_resets USING btree (expires_at);


--
-- Name: idx_password_resets_issued_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_password_resets_issued_at ON public.password_resets USING btree (issued_at);


--
-- Name: idx_password_resets_rate_limit_key; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_password_resets_rate_limit_key ON public.password_resets USING btree (rate_limit_key);


--
-- Name: idx_password_resets_requested_by; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_password_resets_requested_by ON public.password_resets USING btree (requested_by);


--
-- Name: idx_password_resets_should_cleanup; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_password_resets_should_cleanup ON public.password_resets USING btree (should_cleanup);


--
-- Name: idx_password_resets_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_password_resets_status ON public.password_resets USING btree (status);


--
-- Name: idx_password_resets_token; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_password_resets_token ON public.password_resets USING btree (token);


--
-- Name: idx_password_resets_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_password_resets_user_id ON public.password_resets USING btree (user_id);


--
-- Name: idx_payment_gateways_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payment_gateways_deleted_at ON public.payment_gateways USING btree (deleted_at);


--
-- Name: idx_payment_methods_account_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payment_methods_account_id ON public.payment_methods USING btree (account_id);


--
-- Name: idx_payment_methods_active; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payment_methods_active ON public.payment_methods USING btree (account_id, status, is_default) WHERE (status = 'active'::text);


--
-- Name: idx_payment_methods_card_fingerprint; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payment_methods_card_fingerprint ON public.payment_methods USING btree (card_fingerprint);


--
-- Name: idx_payment_methods_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payment_methods_deleted_at ON public.payment_methods USING btree (deleted_at, deleted_at);


--
-- Name: idx_payment_methods_external_payment_method_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payment_methods_external_payment_method_id ON public.payment_methods USING btree (external_payment_method_id);


--
-- Name: idx_payment_methods_is_default; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payment_methods_is_default ON public.payment_methods USING btree (is_default);


--
-- Name: idx_payment_methods_mpesa_phone_number; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payment_methods_mpesa_phone_number ON public.payment_methods USING btree (mpesa_phone_number);


--
-- Name: idx_payment_methods_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payment_methods_status ON public.payment_methods USING btree (status);


--
-- Name: idx_payment_methods_stripe_customer_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payment_methods_stripe_customer_id ON public.payment_methods USING btree (stripe_customer_id);


--
-- Name: idx_payment_methods_stripe_payment_method_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payment_methods_stripe_payment_method_id ON public.payment_methods USING btree (stripe_payment_method_id);


--
-- Name: idx_payment_methods_type; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payment_methods_type ON public.payment_methods USING btree (type);


--
-- Name: idx_payment_records_account_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payment_records_account_id ON public.payment_records USING btree (account_id);


--
-- Name: idx_payment_records_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payment_records_deleted_at ON public.payment_records USING btree (deleted_at, deleted_at);


--
-- Name: idx_payment_records_event_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payment_records_event_id ON public.payment_records USING btree (event_id);


--
-- Name: idx_payment_records_external_transaction_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payment_records_external_transaction_id ON public.payment_records USING btree (external_transaction_id);


--
-- Name: idx_payment_records_lookup; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payment_records_lookup ON public.payment_records USING btree (type, status, initiated_at);


--
-- Name: idx_payment_records_order_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payment_records_order_id ON public.payment_records USING btree (order_id);


--
-- Name: idx_payment_records_organizer_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payment_records_organizer_id ON public.payment_records USING btree (organizer_id);


--
-- Name: idx_payment_records_payment_gateway_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payment_records_payment_gateway_id ON public.payment_records USING btree (payment_gateway_id);


--
-- Name: idx_payment_records_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payment_records_status ON public.payment_records USING btree (status);


--
-- Name: idx_payment_records_type; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payment_records_type ON public.payment_records USING btree (type);


--
-- Name: idx_payment_transactions_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payment_transactions_deleted_at ON public.payment_transactions USING btree (deleted_at);


--
-- Name: idx_payment_transactions_external_transaction_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payment_transactions_external_transaction_id ON public.payment_transactions USING btree (external_transaction_id);


--
-- Name: idx_payment_transactions_order_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payment_transactions_order_id ON public.payment_transactions USING btree (order_id);


--
-- Name: idx_payment_transactions_organizer_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payment_transactions_organizer_id ON public.payment_transactions USING btree (organizer_id);


--
-- Name: idx_payment_transactions_payment_gateway_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payment_transactions_payment_gateway_id ON public.payment_transactions USING btree (payment_gateway_id);


--
-- Name: idx_payment_transactions_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payment_transactions_status ON public.payment_transactions USING btree (status);


--
-- Name: idx_payment_transactions_type; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payment_transactions_type ON public.payment_transactions USING btree (type);


--
-- Name: idx_payout_accounts_account_number; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payout_accounts_account_number ON public.payout_accounts USING btree (account_number);


--
-- Name: idx_payout_accounts_account_type; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payout_accounts_account_type ON public.payout_accounts USING btree (account_type);


--
-- Name: idx_payout_accounts_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payout_accounts_deleted_at ON public.payout_accounts USING btree (deleted_at, deleted_at);


--
-- Name: idx_payout_accounts_external_account_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payout_accounts_external_account_id ON public.payout_accounts USING btree (external_account_id);


--
-- Name: idx_payout_accounts_is_default; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payout_accounts_is_default ON public.payout_accounts USING btree (is_default);


--
-- Name: idx_payout_accounts_is_suspicious_activity; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payout_accounts_is_suspicious_activity ON public.payout_accounts USING btree (is_suspicious_activity);


--
-- Name: idx_payout_accounts_is_verified; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payout_accounts_is_verified ON public.payout_accounts USING btree (is_verified);


--
-- Name: idx_payout_accounts_mobile_phone_number; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payout_accounts_mobile_phone_number ON public.payout_accounts USING btree (mobile_phone_number);


--
-- Name: idx_payout_accounts_organizer_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payout_accounts_organizer_id ON public.payout_accounts USING btree (organizer_id);


--
-- Name: idx_payout_accounts_paypal_email; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payout_accounts_paypal_email ON public.payout_accounts USING btree (paypal_email);


--
-- Name: idx_payout_accounts_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payout_accounts_status ON public.payout_accounts USING btree (status);


--
-- Name: idx_payout_accounts_stripe_account_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payout_accounts_stripe_account_id ON public.payout_accounts USING btree (stripe_account_id);


--
-- Name: idx_payout_accounts_verification_token; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payout_accounts_verification_token ON public.payout_accounts USING btree (verification_token);


--
-- Name: idx_payout_accounts_verified; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payout_accounts_verified ON public.payout_accounts USING btree (organizer_id, status, is_verified) WHERE (status = 'verified'::text);


--
-- Name: idx_promo_usage; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_promo_usage ON public.promotion_usages USING btree (promotion_id);


--
-- Name: idx_promotion_rules_is_active; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_promotion_rules_is_active ON public.promotion_rules USING btree (is_active);


--
-- Name: idx_promotion_rules_promotion_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_promotion_rules_promotion_id ON public.promotion_rules USING btree (promotion_id);


--
-- Name: idx_promotion_rules_rule_type; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_promotion_rules_rule_type ON public.promotion_rules USING btree (rule_type);


--
-- Name: idx_promotion_usages_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_promotion_usages_deleted_at ON public.promotion_usages USING btree (deleted_at);


--
-- Name: idx_promotion_usages_used_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_promotion_usages_used_at ON public.promotion_usages USING btree (used_at);


--
-- Name: idx_promotions_active; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_promotions_active ON public.promotions USING btree (status, start_date, end_date) WHERE (status = 'active'::text);


--
-- Name: idx_promotions_code; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_promotions_code ON public.promotions USING btree (code);


--
-- Name: idx_promotions_created_by; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_promotions_created_by ON public.promotions USING btree (created_by);


--
-- Name: idx_promotions_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_promotions_deleted_at ON public.promotions USING btree (deleted_at, deleted_at);


--
-- Name: idx_promotions_early_bird_cutoff; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_promotions_early_bird_cutoff ON public.promotions USING btree (early_bird_cutoff);


--
-- Name: idx_promotions_end_date; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_promotions_end_date ON public.promotions USING btree (end_date);


--
-- Name: idx_promotions_event_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_promotions_event_id ON public.promotions USING btree (event_id);


--
-- Name: idx_promotions_is_unlimited; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_promotions_is_unlimited ON public.promotions USING btree (is_unlimited);


--
-- Name: idx_promotions_organizer_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_promotions_organizer_id ON public.promotions USING btree (organizer_id);


--
-- Name: idx_promotions_precomputed_active; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_promotions_precomputed_active ON public.promotions USING btree (precomputed_active);


--
-- Name: idx_promotions_start_date; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_promotions_start_date ON public.promotions USING btree (start_date);


--
-- Name: idx_promotions_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_promotions_status ON public.promotions USING btree (status);


--
-- Name: idx_promotions_type; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_promotions_type ON public.promotions USING btree (type);


--
-- Name: idx_promotions_usage; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_promotions_usage ON public.promotions USING btree (usage_count, usage_limit) WHERE (usage_limit IS NOT NULL);


--
-- Name: idx_promotions_usage_count; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_promotions_usage_count ON public.promotions USING btree (usage_count);


--
-- Name: idx_recovery_codes_code_hash; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_recovery_codes_code_hash ON public.recovery_codes USING btree (code_hash);


--
-- Name: idx_recovery_codes_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_recovery_codes_deleted_at ON public.recovery_codes USING btree (deleted_at);


--
-- Name: idx_recovery_codes_two_factor_auth_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_recovery_codes_two_factor_auth_id ON public.recovery_codes USING btree (two_factor_auth_id);


--
-- Name: idx_recovery_codes_unused; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_recovery_codes_unused ON public.recovery_codes USING btree (two_factor_auth_id, used) WHERE (used = false);


--
-- Name: idx_recovery_codes_used; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_recovery_codes_used ON public.recovery_codes USING btree (used);


--
-- Name: idx_refund_line_items_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_refund_line_items_deleted_at ON public.refund_line_items USING btree (deleted_at, deleted_at);


--
-- Name: idx_refund_line_items_lookup; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_refund_line_items_lookup ON public.refund_line_items USING btree (refund_record_id, order_item_id);


--
-- Name: idx_refund_line_items_order_item_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_refund_line_items_order_item_id ON public.refund_line_items USING btree (order_item_id);


--
-- Name: idx_refund_line_items_refund_record_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_refund_line_items_refund_record_id ON public.refund_line_items USING btree (refund_record_id);


--
-- Name: idx_refund_line_items_ticket; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_refund_line_items_ticket ON public.refund_line_items USING btree (ticket_id) WHERE (ticket_id IS NOT NULL);


--
-- Name: idx_refund_line_items_ticket_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_refund_line_items_ticket_id ON public.refund_line_items USING btree (ticket_id);


--
-- Name: idx_refund_records_account_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_refund_records_account_id ON public.refund_records USING btree (account_id);


--
-- Name: idx_refund_records_approved_by; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_refund_records_approved_by ON public.refund_records USING btree (approved_by);


--
-- Name: idx_refund_records_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_refund_records_deleted_at ON public.refund_records USING btree (deleted_at, deleted_at);


--
-- Name: idx_refund_records_event_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_refund_records_event_id ON public.refund_records USING btree (event_id);


--
-- Name: idx_refund_records_external_refund_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_refund_records_external_refund_id ON public.refund_records USING btree (external_refund_id);


--
-- Name: idx_refund_records_order_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_refund_records_order_id ON public.refund_records USING btree (order_id);


--
-- Name: idx_refund_records_organizer_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_refund_records_organizer_id ON public.refund_records USING btree (organizer_id);


--
-- Name: idx_refund_records_payment_gateway_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_refund_records_payment_gateway_id ON public.refund_records USING btree (payment_gateway_id);


--
-- Name: idx_refund_records_refund_number; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_refund_records_refund_number ON public.refund_records USING btree (refund_number);


--
-- Name: idx_refund_records_refund_reason; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_refund_records_refund_reason ON public.refund_records USING btree (refund_reason);


--
-- Name: idx_refund_records_refund_type; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_refund_records_refund_type ON public.refund_records USING btree (refund_type);


--
-- Name: idx_refund_records_requested_by; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_refund_records_requested_by ON public.refund_records USING btree (requested_by);


--
-- Name: idx_refund_records_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_refund_records_status ON public.refund_records USING btree (status);


--
-- Name: idx_reserved_tickets_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_reserved_tickets_deleted_at ON public.reserved_tickets USING btree (deleted_at);


--
-- Name: idx_reserved_tickets_event_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_reserved_tickets_event_id ON public.reserved_tickets USING btree (event_id);


--
-- Name: idx_reserved_tickets_ticket_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_reserved_tickets_ticket_id ON public.reserved_tickets USING btree (ticket_id);


--
-- Name: idx_reset_configurations_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_reset_configurations_deleted_at ON public.reset_configurations USING btree (deleted_at);


--
-- Name: idx_reset_configurations_last_modified_by; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_reset_configurations_last_modified_by ON public.reset_configurations USING btree (last_modified_by);


--
-- Name: idx_security_events; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_security_events ON public.security_metrics USING btree (event_type, "timestamp");


--
-- Name: idx_security_metrics_account_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_security_metrics_account_id ON public.security_metrics USING btree (account_id);


--
-- Name: idx_security_metrics_country; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_security_metrics_country ON public.security_metrics USING btree (country);


--
-- Name: idx_security_metrics_date; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_security_metrics_date ON public.security_metrics USING btree (date);


--
-- Name: idx_security_metrics_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_security_metrics_deleted_at ON public.security_metrics USING btree (deleted_at);


--
-- Name: idx_security_metrics_ip_address; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_security_metrics_ip_address ON public.security_metrics USING btree (ip_address);


--
-- Name: idx_security_metrics_is_resolved; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_security_metrics_is_resolved ON public.security_metrics USING btree (is_resolved);


--
-- Name: idx_security_metrics_resolved_by; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_security_metrics_resolved_by ON public.security_metrics USING btree (resolved_by);


--
-- Name: idx_security_metrics_severity; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_security_metrics_severity ON public.security_metrics USING btree (severity);


--
-- Name: idx_security_metrics_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_security_metrics_user_id ON public.security_metrics USING btree (user_id);


--
-- Name: idx_settlement_items_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_settlement_items_deleted_at ON public.settlement_items USING btree (deleted_at);


--
-- Name: idx_settlement_items_event_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_settlement_items_event_id ON public.settlement_items USING btree (event_id);


--
-- Name: idx_settlement_items_external_transaction_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_settlement_items_external_transaction_id ON public.settlement_items USING btree (external_transaction_id);


--
-- Name: idx_settlement_items_organizer_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_settlement_items_organizer_id ON public.settlement_items USING btree (organizer_id);


--
-- Name: idx_settlement_items_settlement_record_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_settlement_items_settlement_record_id ON public.settlement_items USING btree (settlement_record_id);


--
-- Name: idx_settlement_items_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_settlement_items_status ON public.settlement_items USING btree (status);


--
-- Name: idx_settlement_lookup; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_settlement_lookup ON public.settlement_records USING btree (status, earliest_payout_date);


--
-- Name: idx_settlement_records_approved_by; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_settlement_records_approved_by ON public.settlement_records USING btree (approved_by);


--
-- Name: idx_settlement_records_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_settlement_records_deleted_at ON public.settlement_records USING btree (deleted_at, deleted_at);


--
-- Name: idx_settlement_records_event_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_settlement_records_event_id ON public.settlement_records USING btree (event_id);


--
-- Name: idx_settlement_records_external_batch_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_settlement_records_external_batch_id ON public.settlement_records USING btree (external_batch_id);


--
-- Name: idx_settlement_records_initiated_by; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_settlement_records_initiated_by ON public.settlement_records USING btree (initiated_by);


--
-- Name: idx_settlement_records_payment_gateway_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_settlement_records_payment_gateway_id ON public.settlement_records USING btree (payment_gateway_id);


--
-- Name: idx_settlement_records_period_end_date; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_settlement_records_period_end_date ON public.settlement_records USING btree (period_end_date);


--
-- Name: idx_settlement_records_period_start_date; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_settlement_records_period_start_date ON public.settlement_records USING btree (period_start_date);


--
-- Name: idx_settlement_records_settlement_batch_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_settlement_records_settlement_batch_id ON public.settlement_records USING btree (settlement_batch_id);


--
-- Name: idx_settlement_records_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_settlement_records_status ON public.settlement_records USING btree (status);


--
-- Name: idx_support_ticket_comments_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_support_ticket_comments_deleted_at ON public.support_ticket_comments USING btree (deleted_at);


--
-- Name: idx_support_ticket_comments_ticket_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_support_ticket_comments_ticket_id ON public.support_ticket_comments USING btree (ticket_id);


--
-- Name: idx_support_ticket_comments_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_support_ticket_comments_user_id ON public.support_ticket_comments USING btree (user_id);


--
-- Name: idx_support_tickets_assigned_to_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_support_tickets_assigned_to_id ON public.support_tickets USING btree (assigned_to_id);


--
-- Name: idx_support_tickets_category; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_support_tickets_category ON public.support_tickets USING btree (category);


--
-- Name: idx_support_tickets_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_support_tickets_deleted_at ON public.support_tickets USING btree (deleted_at);


--
-- Name: idx_support_tickets_email; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_support_tickets_email ON public.support_tickets USING btree (email);


--
-- Name: idx_support_tickets_event_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_support_tickets_event_id ON public.support_tickets USING btree (event_id);


--
-- Name: idx_support_tickets_order_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_support_tickets_order_id ON public.support_tickets USING btree (order_id);


--
-- Name: idx_support_tickets_organizer_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_support_tickets_organizer_id ON public.support_tickets USING btree (organizer_id);


--
-- Name: idx_support_tickets_priority; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_support_tickets_priority ON public.support_tickets USING btree (priority);


--
-- Name: idx_support_tickets_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_support_tickets_status ON public.support_tickets USING btree (status);


--
-- Name: idx_support_tickets_ticket_number; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_support_tickets_ticket_number ON public.support_tickets USING btree (ticket_number);


--
-- Name: idx_support_tickets_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_support_tickets_user_id ON public.support_tickets USING btree (user_id);


--
-- Name: idx_system_metrics_account_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_system_metrics_account_id ON public.system_metrics USING btree (account_id);


--
-- Name: idx_system_metrics_city; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_system_metrics_city ON public.system_metrics USING btree (city);


--
-- Name: idx_system_metrics_country; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_system_metrics_country ON public.system_metrics USING btree (country);


--
-- Name: idx_system_metrics_date; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_system_metrics_date ON public.system_metrics USING btree (date);


--
-- Name: idx_system_metrics_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_system_metrics_deleted_at ON public.system_metrics USING btree (deleted_at);


--
-- Name: idx_system_metrics_event_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_system_metrics_event_id ON public.system_metrics USING btree (event_id);


--
-- Name: idx_system_metrics_organizer_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_system_metrics_organizer_id ON public.system_metrics USING btree (organizer_id);


--
-- Name: idx_system_metrics_region; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_system_metrics_region ON public.system_metrics USING btree (region);


--
-- Name: idx_ticket_classes_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_ticket_classes_deleted_at ON public.ticket_classes USING btree (deleted_at);


--
-- Name: idx_ticket_classes_event_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_ticket_classes_event_id ON public.ticket_classes USING btree (event_id);


--
-- Name: idx_ticket_orders_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_ticket_orders_deleted_at ON public.ticket_orders USING btree (deleted_at);


--
-- Name: idx_ticket_orders_order_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_ticket_orders_order_id ON public.ticket_orders USING btree (order_id);


--
-- Name: idx_ticket_orders_ticket_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_ticket_orders_ticket_id ON public.ticket_orders USING btree (ticket_id);


--
-- Name: idx_ticket_transfer_histories_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_ticket_transfer_histories_deleted_at ON public.ticket_transfer_histories USING btree (deleted_at);


--
-- Name: idx_ticket_transfer_histories_ticket_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_ticket_transfer_histories_ticket_id ON public.ticket_transfer_histories USING btree (ticket_id);


--
-- Name: idx_tickets_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_tickets_deleted_at ON public.tickets USING btree (deleted_at);


--
-- Name: idx_tickets_order_item_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_tickets_order_item_id ON public.tickets USING btree (order_item_id);


--
-- Name: idx_timezones_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_timezones_deleted_at ON public.timezones USING btree (deleted_at);


--
-- Name: idx_timezones_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_timezones_name ON public.timezones USING btree (name);


--
-- Name: idx_two_factor_attempts_attempted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_two_factor_attempts_attempted_at ON public.two_factor_attempts USING btree (attempted_at);


--
-- Name: idx_two_factor_attempts_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_two_factor_attempts_deleted_at ON public.two_factor_attempts USING btree (deleted_at);


--
-- Name: idx_two_factor_attempts_success; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_two_factor_attempts_success ON public.two_factor_attempts USING btree (success);


--
-- Name: idx_two_factor_attempts_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_two_factor_attempts_user_id ON public.two_factor_attempts USING btree (user_id);


--
-- Name: idx_two_factor_auths_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_two_factor_auths_deleted_at ON public.two_factor_auths USING btree (deleted_at);


--
-- Name: idx_two_factor_auths_enabled; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_two_factor_auths_enabled ON public.two_factor_auths USING btree (enabled);


--
-- Name: idx_two_factor_auths_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_two_factor_auths_user_id ON public.two_factor_auths USING btree (user_id, user_id);


--
-- Name: idx_two_factor_auths_verified_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_two_factor_auths_verified_at ON public.two_factor_auths USING btree (verified_at);


--
-- Name: idx_two_factor_sessions_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_two_factor_sessions_deleted_at ON public.two_factor_sessions USING btree (deleted_at);


--
-- Name: idx_two_factor_sessions_expires_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_two_factor_sessions_expires_at ON public.two_factor_sessions USING btree (expires_at);


--
-- Name: idx_two_factor_sessions_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_two_factor_sessions_user_id ON public.two_factor_sessions USING btree (user_id);


--
-- Name: idx_user_account; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_user_account ON public.users USING btree (account_id);


--
-- Name: idx_user_engagement; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_user_engagement ON public.user_engagement_metrics USING btree (account_id, date);


--
-- Name: idx_user_engagement_metrics_campaign_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_user_engagement_metrics_campaign_id ON public.user_engagement_metrics USING btree (campaign_id);


--
-- Name: idx_user_engagement_metrics_country; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_user_engagement_metrics_country ON public.user_engagement_metrics USING btree (country);


--
-- Name: idx_user_engagement_metrics_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_user_engagement_metrics_deleted_at ON public.user_engagement_metrics USING btree (deleted_at);


--
-- Name: idx_user_engagement_metrics_referrer_source; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_user_engagement_metrics_referrer_source ON public.user_engagement_metrics USING btree (referrer_source);


--
-- Name: idx_user_usage; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_user_usage ON public.promotion_usages USING btree (account_id);


--
-- Name: idx_users_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_deleted_at ON public.users USING btree (deleted_at);


--
-- Name: idx_users_email; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_users_email ON public.users USING btree (email);


--
-- Name: idx_users_email_verified; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_email_verified ON public.users USING btree (email_verified);


--
-- Name: idx_users_email_verified_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_email_verified_at ON public.users USING btree (email_verified_at);


--
-- Name: idx_users_phone; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_users_phone ON public.users USING btree (phone);


--
-- Name: idx_users_username; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_users_username ON public.users USING btree (username);


--
-- Name: idx_venues_city; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_venues_city ON public.venues USING btree (city);


--
-- Name: idx_venues_country; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_venues_country ON public.venues USING btree (country);


--
-- Name: idx_venues_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_venues_deleted_at ON public.venues USING btree (deleted_at);


--
-- Name: idx_venues_venue_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_venues_venue_name ON public.venues USING btree (venue_name);


--
-- Name: idx_venues_venue_type; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_venues_venue_type ON public.venues USING btree (venue_type);


--
-- Name: idx_waitlist_entries_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_waitlist_entries_deleted_at ON public.waitlist_entries USING btree (deleted_at);


--
-- Name: idx_waitlist_entries_email; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_waitlist_entries_email ON public.waitlist_entries USING btree (email);


--
-- Name: idx_waitlist_entries_event_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_waitlist_entries_event_id ON public.waitlist_entries USING btree (event_id);


--
-- Name: idx_waitlist_entries_session_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_waitlist_entries_session_id ON public.waitlist_entries USING btree (session_id);


--
-- Name: idx_waitlist_entries_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_waitlist_entries_status ON public.waitlist_entries USING btree (status);


--
-- Name: idx_waitlist_entries_ticket_class_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_waitlist_entries_ticket_class_id ON public.waitlist_entries USING btree (ticket_class_id);


--
-- Name: idx_waitlist_entries_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_waitlist_entries_user_id ON public.waitlist_entries USING btree (user_id);


--
-- Name: idx_webhook_logs_account_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_webhook_logs_account_id ON public.webhook_logs USING btree (account_id);


--
-- Name: idx_webhook_logs_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_webhook_logs_deleted_at ON public.webhook_logs USING btree (deleted_at, deleted_at);


--
-- Name: idx_webhook_logs_event_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_webhook_logs_event_id ON public.webhook_logs USING btree (event_id);


--
-- Name: idx_webhook_logs_event_type; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_webhook_logs_event_type ON public.webhook_logs USING btree (event_type);


--
-- Name: idx_webhook_logs_events; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_webhook_logs_events ON public.webhook_logs USING btree (event_id, event_type, created_at);


--
-- Name: idx_webhook_logs_external_reference; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_webhook_logs_external_reference ON public.webhook_logs USING btree (external_reference);


--
-- Name: idx_webhook_logs_external_transaction_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_webhook_logs_external_transaction_id ON public.webhook_logs USING btree (external_transaction_id);


--
-- Name: idx_webhook_logs_idempotency_key; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_webhook_logs_idempotency_key ON public.webhook_logs USING btree (idempotency_key);


--
-- Name: idx_webhook_logs_ip_address; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_webhook_logs_ip_address ON public.webhook_logs USING btree (ip_address);


--
-- Name: idx_webhook_logs_is_duplicate; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_webhook_logs_is_duplicate ON public.webhook_logs USING btree (is_duplicate);


--
-- Name: idx_webhook_logs_order_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_webhook_logs_order_id ON public.webhook_logs USING btree (order_id);


--
-- Name: idx_webhook_logs_organizer_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_webhook_logs_organizer_id ON public.webhook_logs USING btree (organizer_id);


--
-- Name: idx_webhook_logs_payment_record_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_webhook_logs_payment_record_id ON public.webhook_logs USING btree (payment_record_id);


--
-- Name: idx_webhook_logs_payment_transaction_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_webhook_logs_payment_transaction_id ON public.webhook_logs USING btree (payment_transaction_id);


--
-- Name: idx_webhook_logs_processing; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_webhook_logs_processing ON public.webhook_logs USING btree (provider, status, created_at);


--
-- Name: idx_webhook_logs_provider; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_webhook_logs_provider ON public.webhook_logs USING btree (provider);


--
-- Name: idx_webhook_logs_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_webhook_logs_status ON public.webhook_logs USING btree (status);


--
-- Name: idx_webhook_logs_success; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_webhook_logs_success ON public.webhook_logs USING btree (success);


--
-- Name: account_activities fk_account_activities_account; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.account_activities
    ADD CONSTRAINT fk_account_activities_account FOREIGN KEY (account_id) REFERENCES public.accounts(id);


--
-- Name: account_activities fk_account_activities_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.account_activities
    ADD CONSTRAINT fk_account_activities_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: account_payment_gateways fk_account_payment_gateways_account; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.account_payment_gateways
    ADD CONSTRAINT fk_account_payment_gateways_account FOREIGN KEY (account_id) REFERENCES public.accounts(id);


--
-- Name: account_payment_gateways fk_account_payment_gateways_payment_gateway; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.account_payment_gateways
    ADD CONSTRAINT fk_account_payment_gateways_payment_gateway FOREIGN KEY (payment_gateway_id) REFERENCES public.payment_gateways(id);


--
-- Name: accounts fk_accounts_payment_gateway; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.accounts
    ADD CONSTRAINT fk_accounts_payment_gateway FOREIGN KEY (payment_gateway_id) REFERENCES public.payment_gateways(id);


--
-- Name: attendees fk_attendees_account; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.attendees
    ADD CONSTRAINT fk_attendees_account FOREIGN KEY (account_id) REFERENCES public.accounts(id);


--
-- Name: attendees fk_attendees_event; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.attendees
    ADD CONSTRAINT fk_attendees_event FOREIGN KEY (event_id) REFERENCES public.events(id);


--
-- Name: attendees fk_attendees_order; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.attendees
    ADD CONSTRAINT fk_attendees_order FOREIGN KEY (order_id) REFERENCES public.orders(id);


--
-- Name: attendees fk_attendees_ticket; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.attendees
    ADD CONSTRAINT fk_attendees_ticket FOREIGN KEY (ticket_id) REFERENCES public.tickets(id);


--
-- Name: email_verifications fk_email_verifications_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.email_verifications
    ADD CONSTRAINT fk_email_verifications_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: event_images fk_event_images_account; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.event_images
    ADD CONSTRAINT fk_event_images_account FOREIGN KEY (account_id) REFERENCES public.accounts(id);


--
-- Name: event_images fk_event_images_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.event_images
    ADD CONSTRAINT fk_event_images_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: event_metrics fk_event_metrics_event; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.event_metrics
    ADD CONSTRAINT fk_event_metrics_event FOREIGN KEY (event_id) REFERENCES public.events(id);


--
-- Name: event_stats fk_event_stats_event; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.event_stats
    ADD CONSTRAINT fk_event_stats_event FOREIGN KEY (event_id) REFERENCES public.events(id);


--
-- Name: event_venues fk_event_venues_event; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.event_venues
    ADD CONSTRAINT fk_event_venues_event FOREIGN KEY (event_id) REFERENCES public.events(id);


--
-- Name: event_venues fk_event_venues_venue; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.event_venues
    ADD CONSTRAINT fk_event_venues_venue FOREIGN KEY (venue_id) REFERENCES public.venues(id);


--
-- Name: events fk_events_account; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.events
    ADD CONSTRAINT fk_events_account FOREIGN KEY (account_id) REFERENCES public.accounts(id);


--
-- Name: event_images fk_events_event_images; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.event_images
    ADD CONSTRAINT fk_events_event_images FOREIGN KEY (event_id) REFERENCES public.events(id);


--
-- Name: events fk_events_organizer; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.events
    ADD CONSTRAINT fk_events_organizer FOREIGN KEY (organizer_id) REFERENCES public.organizers(id);


--
-- Name: login_history fk_login_history_account; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.login_history
    ADD CONSTRAINT fk_login_history_account FOREIGN KEY (account_id) REFERENCES public.accounts(id);


--
-- Name: login_history fk_login_history_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.login_history
    ADD CONSTRAINT fk_login_history_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: notification_preferences fk_notification_preferences_account; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notification_preferences
    ADD CONSTRAINT fk_notification_preferences_account FOREIGN KEY (account_id) REFERENCES public.accounts(id);


--
-- Name: tickets fk_order_items_generated_tickets; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tickets
    ADD CONSTRAINT fk_order_items_generated_tickets FOREIGN KEY (order_item_id) REFERENCES public.order_items(id);


--
-- Name: order_items fk_order_items_ticket_class; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT fk_order_items_ticket_class FOREIGN KEY (ticket_class_id) REFERENCES public.ticket_classes(id);


--
-- Name: orders fk_orders_account; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT fk_orders_account FOREIGN KEY (account_id) REFERENCES public.accounts(id);


--
-- Name: orders fk_orders_event; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT fk_orders_event FOREIGN KEY (event_id) REFERENCES public.events(id);


--
-- Name: order_items fk_orders_order_items; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT fk_orders_order_items FOREIGN KEY (order_id) REFERENCES public.orders(id);


--
-- Name: orders fk_orders_payment_gateway; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT fk_orders_payment_gateway FOREIGN KEY (payment_gateway_id) REFERENCES public.payment_gateways(id);


--
-- Name: payment_records fk_orders_payment_records; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payment_records
    ADD CONSTRAINT fk_orders_payment_records FOREIGN KEY (order_id) REFERENCES public.orders(id);


--
-- Name: organizers fk_organizers_account; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.organizers
    ADD CONSTRAINT fk_organizers_account FOREIGN KEY (account_id) REFERENCES public.accounts(id);


--
-- Name: organizers fk_organizers_payment_gateway; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.organizers
    ADD CONSTRAINT fk_organizers_payment_gateway FOREIGN KEY (payment_gateway_id) REFERENCES public.payment_gateways(id);


--
-- Name: password_reset_attempts fk_password_reset_attempts_password_reset; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.password_reset_attempts
    ADD CONSTRAINT fk_password_reset_attempts_password_reset FOREIGN KEY (password_reset_id) REFERENCES public.password_resets(id);


--
-- Name: password_resets fk_password_resets_account; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.password_resets
    ADD CONSTRAINT fk_password_resets_account FOREIGN KEY (account_id) REFERENCES public.accounts(id);


--
-- Name: password_resets fk_password_resets_approved_by_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.password_resets
    ADD CONSTRAINT fk_password_resets_approved_by_user FOREIGN KEY (approved_by) REFERENCES public.users(id);


--
-- Name: password_resets fk_password_resets_requested_by_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.password_resets
    ADD CONSTRAINT fk_password_resets_requested_by_user FOREIGN KEY (requested_by) REFERENCES public.users(id);


--
-- Name: password_resets fk_password_resets_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.password_resets
    ADD CONSTRAINT fk_password_resets_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: payment_methods fk_payment_methods_account; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payment_methods
    ADD CONSTRAINT fk_payment_methods_account FOREIGN KEY (account_id) REFERENCES public.accounts(id);


--
-- Name: payment_records fk_payment_records_account; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payment_records
    ADD CONSTRAINT fk_payment_records_account FOREIGN KEY (account_id) REFERENCES public.accounts(id);


--
-- Name: payment_records fk_payment_records_child_records; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payment_records
    ADD CONSTRAINT fk_payment_records_child_records FOREIGN KEY (parent_record_id) REFERENCES public.payment_records(id);


--
-- Name: payment_records fk_payment_records_event; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payment_records
    ADD CONSTRAINT fk_payment_records_event FOREIGN KEY (event_id) REFERENCES public.events(id);


--
-- Name: payment_records fk_payment_records_organizer; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payment_records
    ADD CONSTRAINT fk_payment_records_organizer FOREIGN KEY (organizer_id) REFERENCES public.organizers(id);


--
-- Name: payment_records fk_payment_records_payment_gateway; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payment_records
    ADD CONSTRAINT fk_payment_records_payment_gateway FOREIGN KEY (payment_gateway_id) REFERENCES public.payment_gateways(id);


--
-- Name: payment_transactions fk_payment_transactions_order; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payment_transactions
    ADD CONSTRAINT fk_payment_transactions_order FOREIGN KEY (order_id) REFERENCES public.orders(id);


--
-- Name: payment_transactions fk_payment_transactions_organizer; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payment_transactions
    ADD CONSTRAINT fk_payment_transactions_organizer FOREIGN KEY (organizer_id) REFERENCES public.organizers(id);


--
-- Name: payment_transactions fk_payment_transactions_parent_transaction; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payment_transactions
    ADD CONSTRAINT fk_payment_transactions_parent_transaction FOREIGN KEY (parent_transaction_id) REFERENCES public.payment_transactions(id);


--
-- Name: payment_transactions fk_payment_transactions_payment_gateway; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payment_transactions
    ADD CONSTRAINT fk_payment_transactions_payment_gateway FOREIGN KEY (payment_gateway_id) REFERENCES public.payment_gateways(id);


--
-- Name: payout_accounts fk_payout_accounts_organizer; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payout_accounts
    ADD CONSTRAINT fk_payout_accounts_organizer FOREIGN KEY (organizer_id) REFERENCES public.organizers(id);


--
-- Name: promotion_rules fk_promotion_rules_promotion; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.promotion_rules
    ADD CONSTRAINT fk_promotion_rules_promotion FOREIGN KEY (promotion_id) REFERENCES public.promotions(id);


--
-- Name: promotion_usages fk_promotion_usages_account; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.promotion_usages
    ADD CONSTRAINT fk_promotion_usages_account FOREIGN KEY (account_id) REFERENCES public.accounts(id);


--
-- Name: promotion_usages fk_promotion_usages_order; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.promotion_usages
    ADD CONSTRAINT fk_promotion_usages_order FOREIGN KEY (order_id) REFERENCES public.orders(id);


--
-- Name: promotion_usages fk_promotion_usages_promotion; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.promotion_usages
    ADD CONSTRAINT fk_promotion_usages_promotion FOREIGN KEY (promotion_id) REFERENCES public.promotions(id);


--
-- Name: promotions fk_promotions_created_by_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.promotions
    ADD CONSTRAINT fk_promotions_created_by_user FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: promotions fk_promotions_event; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.promotions
    ADD CONSTRAINT fk_promotions_event FOREIGN KEY (event_id) REFERENCES public.events(id);


--
-- Name: promotions fk_promotions_organizer; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.promotions
    ADD CONSTRAINT fk_promotions_organizer FOREIGN KEY (organizer_id) REFERENCES public.organizers(id);


--
-- Name: recovery_codes fk_recovery_codes_two_factor_auth; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.recovery_codes
    ADD CONSTRAINT fk_recovery_codes_two_factor_auth FOREIGN KEY (two_factor_auth_id) REFERENCES public.two_factor_auths(id);


--
-- Name: refund_line_items fk_refund_line_items_order_item; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.refund_line_items
    ADD CONSTRAINT fk_refund_line_items_order_item FOREIGN KEY (order_item_id) REFERENCES public.order_items(id);


--
-- Name: refund_line_items fk_refund_line_items_ticket; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.refund_line_items
    ADD CONSTRAINT fk_refund_line_items_ticket FOREIGN KEY (ticket_id) REFERENCES public.tickets(id);


--
-- Name: refund_records fk_refund_records_account; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.refund_records
    ADD CONSTRAINT fk_refund_records_account FOREIGN KEY (account_id) REFERENCES public.accounts(id);


--
-- Name: refund_records fk_refund_records_approved_by_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.refund_records
    ADD CONSTRAINT fk_refund_records_approved_by_user FOREIGN KEY (approved_by) REFERENCES public.users(id);


--
-- Name: refund_records fk_refund_records_event; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.refund_records
    ADD CONSTRAINT fk_refund_records_event FOREIGN KEY (event_id) REFERENCES public.events(id);


--
-- Name: refund_records fk_refund_records_order; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.refund_records
    ADD CONSTRAINT fk_refund_records_order FOREIGN KEY (order_id) REFERENCES public.orders(id);


--
-- Name: refund_records fk_refund_records_organizer; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.refund_records
    ADD CONSTRAINT fk_refund_records_organizer FOREIGN KEY (organizer_id) REFERENCES public.organizers(id);


--
-- Name: refund_records fk_refund_records_payment_gateway; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.refund_records
    ADD CONSTRAINT fk_refund_records_payment_gateway FOREIGN KEY (payment_gateway_id) REFERENCES public.payment_gateways(id);


--
-- Name: refund_line_items fk_refund_records_refund_line_items; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.refund_line_items
    ADD CONSTRAINT fk_refund_records_refund_line_items FOREIGN KEY (refund_record_id) REFERENCES public.refund_records(id);


--
-- Name: refund_records fk_refund_records_requested_by_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.refund_records
    ADD CONSTRAINT fk_refund_records_requested_by_user FOREIGN KEY (requested_by) REFERENCES public.users(id);


--
-- Name: reserved_tickets fk_reserved_tickets_event; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.reserved_tickets
    ADD CONSTRAINT fk_reserved_tickets_event FOREIGN KEY (event_id) REFERENCES public.events(id);


--
-- Name: reserved_tickets fk_reserved_tickets_ticket; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.reserved_tickets
    ADD CONSTRAINT fk_reserved_tickets_ticket FOREIGN KEY (ticket_id) REFERENCES public.tickets(id);


--
-- Name: reset_configurations fk_reset_configurations_created_by_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.reset_configurations
    ADD CONSTRAINT fk_reset_configurations_created_by_user FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: reset_configurations fk_reset_configurations_last_modified_by_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.reset_configurations
    ADD CONSTRAINT fk_reset_configurations_last_modified_by_user FOREIGN KEY (last_modified_by) REFERENCES public.users(id);


--
-- Name: security_metrics fk_security_metrics_account; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.security_metrics
    ADD CONSTRAINT fk_security_metrics_account FOREIGN KEY (account_id) REFERENCES public.accounts(id);


--
-- Name: security_metrics fk_security_metrics_resolved_by_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.security_metrics
    ADD CONSTRAINT fk_security_metrics_resolved_by_user FOREIGN KEY (resolved_by) REFERENCES public.users(id);


--
-- Name: security_metrics fk_security_metrics_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.security_metrics
    ADD CONSTRAINT fk_security_metrics_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: settlement_items fk_settlement_items_event; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.settlement_items
    ADD CONSTRAINT fk_settlement_items_event FOREIGN KEY (event_id) REFERENCES public.events(id);


--
-- Name: settlement_items fk_settlement_items_organizer; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.settlement_items
    ADD CONSTRAINT fk_settlement_items_organizer FOREIGN KEY (organizer_id) REFERENCES public.organizers(id);


--
-- Name: settlement_payment_records fk_settlement_payment_records_payment_record; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.settlement_payment_records
    ADD CONSTRAINT fk_settlement_payment_records_payment_record FOREIGN KEY (payment_record_id) REFERENCES public.payment_records(id);


--
-- Name: settlement_payment_records fk_settlement_payment_records_settlement_item; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.settlement_payment_records
    ADD CONSTRAINT fk_settlement_payment_records_settlement_item FOREIGN KEY (settlement_item_id) REFERENCES public.settlement_items(id);


--
-- Name: settlement_records fk_settlement_records_approved_by_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.settlement_records
    ADD CONSTRAINT fk_settlement_records_approved_by_user FOREIGN KEY (approved_by) REFERENCES public.users(id);


--
-- Name: settlement_records fk_settlement_records_event; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.settlement_records
    ADD CONSTRAINT fk_settlement_records_event FOREIGN KEY (event_id) REFERENCES public.events(id);


--
-- Name: settlement_records fk_settlement_records_initiated_by_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.settlement_records
    ADD CONSTRAINT fk_settlement_records_initiated_by_user FOREIGN KEY (initiated_by) REFERENCES public.users(id);


--
-- Name: settlement_records fk_settlement_records_payment_gateway; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.settlement_records
    ADD CONSTRAINT fk_settlement_records_payment_gateway FOREIGN KEY (payment_gateway_id) REFERENCES public.payment_gateways(id);


--
-- Name: settlement_items fk_settlement_records_settlement_items; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.settlement_items
    ADD CONSTRAINT fk_settlement_records_settlement_items FOREIGN KEY (settlement_record_id) REFERENCES public.settlement_records(id);


--
-- Name: support_ticket_comments fk_support_ticket_comments_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.support_ticket_comments
    ADD CONSTRAINT fk_support_ticket_comments_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: support_tickets fk_support_tickets_assigned_to; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.support_tickets
    ADD CONSTRAINT fk_support_tickets_assigned_to FOREIGN KEY (assigned_to_id) REFERENCES public.users(id);


--
-- Name: support_ticket_comments fk_support_tickets_comments; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.support_ticket_comments
    ADD CONSTRAINT fk_support_tickets_comments FOREIGN KEY (ticket_id) REFERENCES public.support_tickets(id);


--
-- Name: support_tickets fk_support_tickets_event; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.support_tickets
    ADD CONSTRAINT fk_support_tickets_event FOREIGN KEY (event_id) REFERENCES public.events(id);


--
-- Name: support_tickets fk_support_tickets_order; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.support_tickets
    ADD CONSTRAINT fk_support_tickets_order FOREIGN KEY (order_id) REFERENCES public.orders(id);


--
-- Name: support_tickets fk_support_tickets_organizer; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.support_tickets
    ADD CONSTRAINT fk_support_tickets_organizer FOREIGN KEY (organizer_id) REFERENCES public.organizers(id);


--
-- Name: support_tickets fk_support_tickets_resolved_by; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.support_tickets
    ADD CONSTRAINT fk_support_tickets_resolved_by FOREIGN KEY (resolved_by_id) REFERENCES public.users(id);


--
-- Name: support_tickets fk_support_tickets_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.support_tickets
    ADD CONSTRAINT fk_support_tickets_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: system_metrics fk_system_metrics_account; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.system_metrics
    ADD CONSTRAINT fk_system_metrics_account FOREIGN KEY (account_id) REFERENCES public.accounts(id);


--
-- Name: system_metrics fk_system_metrics_event; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.system_metrics
    ADD CONSTRAINT fk_system_metrics_event FOREIGN KEY (event_id) REFERENCES public.events(id);


--
-- Name: system_metrics fk_system_metrics_organizer; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.system_metrics
    ADD CONSTRAINT fk_system_metrics_organizer FOREIGN KEY (organizer_id) REFERENCES public.organizers(id);


--
-- Name: ticket_classes fk_ticket_classes_event; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ticket_classes
    ADD CONSTRAINT fk_ticket_classes_event FOREIGN KEY (event_id) REFERENCES public.events(id);


--
-- Name: ticket_orders fk_ticket_orders_order; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ticket_orders
    ADD CONSTRAINT fk_ticket_orders_order FOREIGN KEY (order_id) REFERENCES public.orders(id);


--
-- Name: ticket_orders fk_ticket_orders_ticket; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ticket_orders
    ADD CONSTRAINT fk_ticket_orders_ticket FOREIGN KEY (ticket_id) REFERENCES public.tickets(id);


--
-- Name: ticket_transfer_histories fk_ticket_transfer_histories_ticket; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ticket_transfer_histories
    ADD CONSTRAINT fk_ticket_transfer_histories_ticket FOREIGN KEY (ticket_id) REFERENCES public.tickets(id);


--
-- Name: ticket_transfer_histories fk_tickets_transfer_history; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ticket_transfer_histories
    ADD CONSTRAINT fk_tickets_transfer_history FOREIGN KEY (ticket_id) REFERENCES public.tickets(id);


--
-- Name: two_factor_attempts fk_two_factor_attempts_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.two_factor_attempts
    ADD CONSTRAINT fk_two_factor_attempts_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: two_factor_auths fk_two_factor_auths_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.two_factor_auths
    ADD CONSTRAINT fk_two_factor_auths_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: two_factor_sessions fk_two_factor_sessions_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.two_factor_sessions
    ADD CONSTRAINT fk_two_factor_sessions_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: user_engagement_metrics fk_user_engagement_metrics_account; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_engagement_metrics
    ADD CONSTRAINT fk_user_engagement_metrics_account FOREIGN KEY (account_id) REFERENCES public.accounts(id);


--
-- Name: users fk_users_account; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT fk_users_account FOREIGN KEY (account_id) REFERENCES public.accounts(id);


--
-- Name: waitlist_entries fk_waitlist_entries_event; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.waitlist_entries
    ADD CONSTRAINT fk_waitlist_entries_event FOREIGN KEY (event_id) REFERENCES public.events(id);


--
-- Name: waitlist_entries fk_waitlist_entries_ticket_class; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.waitlist_entries
    ADD CONSTRAINT fk_waitlist_entries_ticket_class FOREIGN KEY (ticket_class_id) REFERENCES public.ticket_classes(id);


--
-- Name: waitlist_entries fk_waitlist_entries_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.waitlist_entries
    ADD CONSTRAINT fk_waitlist_entries_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- PostgreSQL database dump complete
--

\unrestrict euhg8NAxHPDbyzdM1rc8sbuidy5sW9ShGZV2Jgw7fd0BkaQYauvZjJYY55RcxwN

